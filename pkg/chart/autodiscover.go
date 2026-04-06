/*
 * Copyright 2022 IPONWEB
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package chart

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/iponweb/metachart/pkg/helpers"
)

// ---- Kubernetes discovery format types -----

type apiGroupList struct {
	Groups []apiGroup `json:"groups"`
}

type apiGroup struct {
	Name     string            `json:"name"`
	Versions []apiGroupVersion `json:"versions"`
}

type apiGroupVersion struct {
	GroupVersion string `json:"groupVersion"`
	Version      string `json:"version"`
}

type apiResourceList struct {
	GroupVersion string        `json:"groupVersion"`
	Resources    []apiResource `json:"resources"`
}

type apiResource struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// gvk mirrors the x-kubernetes-group-version-kind annotation in a definition.
type gvk struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

// pendingRelated records that a ConversionRule needs its Related map populated
// after all autodiscover sources have been processed (cross-source resolution).
type pendingRelated struct {
	ruleTarget string
	related    []ResourceSelector
}

// discoverResult holds the output of processing one AutodiscoverSource.
type discoverResult struct {
	rules     []ConversionRule
	resources map[string]ResourceDefinition
	pending   []pendingRelated
}

// ---- Autodiscover -----

// Autodiscover processes every AutodiscoverSource in the chart config and
// merges the results into the in-memory config:
//
//   - Each source's Definitions URL is prepended to spec.schema.definitions
//     (deduplicated, autodiscover definitions come before explicit ones).
//   - Discovered ConversionRules are prepended to spec.schema.rules so that
//     manually written rules take effect last and can override them.
//   - Discovered ResourceDefinitions are merged into spec.resources; manual
//     entries with the same key always win.
//   - Related selectors on include rules are resolved after all sources are
//     processed so cross-source references work correctly.
func (c *Chart) Autodiscover() error {
	type srcResult struct {
		result  discoverResult
		defsURL helpers.FilePath
	}

	// Pass 1: process every source independently.
	var srcResults []srcResult
	for _, src := range c.Config.Spec.Autodiscover {
		// Render {{ .version }} before anything else.
		if src.Version != "" {
			var err error
			if src.Discovery, err = renderURLTemplate(src.Discovery, src.Version); err != nil {
				return fmt.Errorf("autodiscover discovery URL: %w", err)
			}
			if src.Definitions, err = renderURLTemplate(src.Definitions, src.Version); err != nil {
				return fmt.Errorf("autodiscover definitions URL: %w", err)
			}
		}

		result, err := processAutodiscoverSource(src)
		if err != nil {
			return err
		}
		srcResults = append(srcResults, srcResult{result: result, defsURL: src.Definitions})
	}

	// Collect unique definition URLs: normal sources first, definitionsLast
	// sources appended after so their types take precedence (last write wins).
	seen := make(map[helpers.FilePath]bool)
	var autodiscoverDefs []helpers.FilePath
	var autodiscoverDefsLast []helpers.FilePath
	for i, sr := range srcResults {
		if seen[sr.defsURL] {
			continue
		}
		seen[sr.defsURL] = true
		if c.Config.Spec.Autodiscover[i].DefinitionsLast {
			autodiscoverDefsLast = append(autodiscoverDefsLast, sr.defsURL)
		} else {
			autodiscoverDefs = append(autodiscoverDefs, sr.defsURL)
		}
	}
	autodiscoverDefs = append(autodiscoverDefs, autodiscoverDefsLast...)

	// Pass 2: resolve pending related selectors.
	// Build a combined resource map (explicit wins, then first autodiscovered
	// source wins) so related selectors can reference any discovered resource.
	hasPending := false
	for _, sr := range srcResults {
		if len(sr.result.pending) > 0 {
			hasPending = true
			break
		}
	}
	if hasPending {
		allResources := make(map[string]ResourceDefinition)
		for k, v := range c.Config.Spec.Resources {
			allResources[k] = v
		}
		for _, sr := range srcResults {
			for k, v := range sr.result.resources {
				if _, exists := allResources[k]; !exists {
					allResources[k] = v
				}
			}
		}

		// For plural-only related selectors, build a plural→latestGV map from
		// the combined resources (they already reflect the winning version).
		relatedPluralLatest := make(map[string]string, len(allResources))
		for plural, resDef := range allResources {
			relatedPluralLatest[plural] = resDef.ApiVersion
		}

		// Index all autodiscovered rules by target so we can patch them.
		rulesByTarget := make(map[string]*ConversionRule)
		for sri := range srcResults {
			for ri := range srcResults[sri].result.rules {
				r := &srcResults[sri].result.rules[ri]
				rulesByTarget[r.Target] = r
			}
		}

		for _, sr := range srcResults {
			for _, pr := range sr.result.pending {
				rule, ok := rulesByTarget[pr.ruleTarget]
				if !ok {
					continue
				}
				if rule.Related == nil {
					rule.Related = make(map[string]string)
				}
				for _, sel := range pr.related {
					for plural, resDef := range allResources {
						if selectorMatches(sel, resDef.ApiVersion, resDef.Kind, plural, relatedPluralLatest) {
							rule.Related[plural] = resDef.JsonSchemaRef
						}
					}
				}
			}
		}
	}

	// Apply results: prepend each source's rules (so manual rules, appended
	// last, take highest precedence) and merge resources.
	if c.Config.Spec.Resources == nil {
		c.Config.Spec.Resources = make(map[string]ResourceDefinition)
	}
	for _, sr := range srcResults {
		c.Config.Spec.Schema.Rules = append(sr.result.rules, c.Config.Spec.Schema.Rules...)
		for key, def := range sr.result.resources {
			if _, exists := c.Config.Spec.Resources[key]; !exists {
				c.Config.Spec.Resources[key] = def
			}
		}
	}

	// Merge order: autodiscover definitions first, then explicit ones.
	c.Config.Spec.Schema.Definitions = append(autodiscoverDefs, c.Config.Spec.Schema.Definitions...)

	return nil
}

func processAutodiscoverSource(src AutodiscoverSource) (discoverResult, error) {
	// Validate selectors: kind and plural are mutually exclusive.
	for _, sel := range append(src.Include, src.Exclude...) {
		if len(sel.Kind) > 0 && len(sel.Plural) > 0 {
			return discoverResult{}, fmt.Errorf(
				"resource selector cannot have both kind and plural set (apiVersion %q)", sel.APIVersion)
		}
	}

	// Fetch the discovery document.
	apisData, err := src.Discovery.Read()
	if err != nil {
		return discoverResult{}, fmt.Errorf("fetch discovery %s: %w", src.Discovery, err)
	}

	// Probe the discovery document kind to distinguish APIGroupList from
	// APIResourceList (e.g. api__v1.json for the core Kubernetes API).
	var kindProbe struct {
		Kind string `json:"kind"`
	}
	_ = json.Unmarshal(apisData, &kindProbe)

	// Fetch and index the definitions file by (group, version, kind).
	defsData, err := src.Definitions.Read()
	if err != nil {
		return discoverResult{}, fmt.Errorf("fetch definitions %s: %w", src.Definitions, err)
	}
	defIndex, err := buildDefinitionIndex(defsData)
	if err != nil {
		return discoverResult{}, fmt.Errorf("index definitions: %w", err)
	}

	// Build a unified list of (group, resourceList) pairs regardless of
	// which discovery format the source uses.
	type groupedList struct {
		group   string
		resList *apiResourceList
	}
	var lists []groupedList

	if kindProbe.Kind == "APIResourceList" {
		// Discovery document is itself a resource list (e.g. api__v1.json
		// for the core Kubernetes API group).
		var resList apiResourceList
		if err := json.Unmarshal(apisData, &resList); err != nil {
			return discoverResult{}, fmt.Errorf("parse APIResourceList: %w", err)
		}
		group, _ := splitGroupVersion(resList.GroupVersion)
		lists = append(lists, groupedList{group: group, resList: &resList})
	} else {
		// Standard APIGroupList document (apis.json).
		var groupList apiGroupList
		if err := json.Unmarshal(apisData, &groupList); err != nil {
			return discoverResult{}, fmt.Errorf("parse apis.json: %w", err)
		}
		for _, group := range groupList.Groups {
			for _, gv := range group.Versions {
				resList, err := fetchResourceList(src.Discovery, gv.GroupVersion)
				if err != nil {
					return discoverResult{}, err
				}
				lists = append(lists, groupedList{group: group.Name, resList: resList})
			}
		}
	}

	// Build the plural→latestGroupVersion map if any selector uses plural-only
	// mode (no apiVersion), so those selectors can resolve the correct version.
	needLatest := false
	for _, sel := range append(src.Include, src.Exclude...) {
		if sel.APIVersion == "" && len(sel.Plural) > 0 {
			needLatest = true
			break
		}
	}

	var pluralLatest map[string]string
	if needLatest {
		type ventry struct {
			groupVersion string
			version      string
		}
		byPlural := map[string][]ventry{}
		for _, gl := range lists {
			_, version := splitGroupVersion(gl.resList.GroupVersion)
			for _, res := range gl.resList.Resources {
				if strings.Contains(res.Name, "/") {
					continue
				}
				byPlural[res.Name] = append(byPlural[res.Name], ventry{
					groupVersion: gl.resList.GroupVersion,
					version:      version,
				})
			}
		}
		pluralLatest = make(map[string]string, len(byPlural))
		for plural, entries := range byPlural {
			sort.Slice(entries, func(i, j int) bool {
				return kubeVersionIsNewer(entries[i].version, entries[j].version)
			})
			pluralLatest[plural] = entries[0].groupVersion
		}
	}

	var result discoverResult
	result.resources = make(map[string]ResourceDefinition)

	for _, gl := range lists {
		_, version := splitGroupVersion(gl.resList.GroupVersion)
		for _, res := range gl.resList.Resources {
			// Skip subresources (e.g. "pods/status").
			if strings.Contains(res.Name, "/") {
				continue
			}

			if !matchesFilters(gl.resList.GroupVersion, res.Kind, res.Name, src.Include, src.Exclude, pluralLatest) {
				continue
			}

			defKey, ok := defIndex[gvk{Group: gl.group, Version: version, Kind: res.Kind}]
			if !ok {
				return discoverResult{}, fmt.Errorf(
					"no definition found for %s/%s (kind %s) — check the definitions URL",
					gl.resList.GroupVersion, res.Name, res.Kind)
			}

			target := "metachart.api." + defKey

			// Start with default properties; find the matching include selector
			// and merge its properties on top (selector properties win).
			props := map[string]string{
				"enabled":  "metachart.interface.boolean",
				"metadata": "metachart.api.meta.v1.ObjectMeta",
			}
			var matchedInclude *ResourceSelector
			for i := range src.Include {
				if selectorMatches(src.Include[i], gl.resList.GroupVersion, res.Kind, res.Name, pluralLatest) {
					matchedInclude = &src.Include[i]
					break
				}
			}
			disallowed := []string{"status", "kind", "apiVersion"}
			if matchedInclude != nil {
				disallowed = append(disallowed, matchedInclude.Disallowed...)
				for k, v := range matchedInclude.Properties {
					props[k] = v
				}
			}
			result.rules = append(result.rules, ConversionRule{
				Source:     &defKey,
				Target:     target,
				Disallowed: &disallowed,
				Properties: &props,
			})

			result.resources[res.Name] = ResourceDefinition{
				ApiVersion:    gl.resList.GroupVersion,
				Kind:          res.Kind,
				JsonSchemaRef: target,
				Template:      true,
				Root:          true,
				Defaults:      true,
			}

			// Schedule related resolution for after all sources are processed.
			if matchedInclude != nil && len(matchedInclude.Related) > 0 {
				result.pending = append(result.pending, pendingRelated{
					ruleTarget: target,
					related:    matchedInclude.Related,
				})
			}
		}
	}

	return result, nil
}

// buildDefinitionIndex parses a _definitions.json payload and returns a map
// from (group, version, kind) to definition key, using the
// x-kubernetes-group-version-kind annotation present on each root resource
// definition.
func buildDefinitionIndex(data []byte) (map[gvk]string, error) {
	var wrapper struct {
		Definitions map[string]json.RawMessage `json:"definitions"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	index := make(map[gvk]string, len(wrapper.Definitions))

	for key, raw := range wrapper.Definitions {
		var entry struct {
			GVKs []gvk `json:"x-kubernetes-group-version-kind"`
		}
		if err := json.Unmarshal(raw, &entry); err != nil {
			continue
		}
		for _, g := range entry.GVKs {
			index[g] = key
		}
	}

	return index, nil
}

// fetchResourceList derives the per-group-version resource list URL from the
// apis.json URL and fetches it.
//
// Convention: given https://…/discovery/apis.json and groupVersion
// "gateway.networking.k8s.io/v1", the resource list lives at
// https://…/discovery/apis__gateway.networking.k8s.io__v1.json
func fetchResourceList(apisURL helpers.FilePath, groupVersion string) (*apiResourceList, error) {
	parts := strings.SplitN(groupVersion, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid groupVersion %q", groupVersion)
	}
	group, version := parts[0], parts[1]

	s := string(apisURL)
	baseURL := s[:strings.LastIndex(s, "/")+1]
	listURL := helpers.FilePath(fmt.Sprintf("%sapis__%s__%s.json", baseURL, group, version))

	listData, err := listURL.Read()
	if err != nil {
		return nil, fmt.Errorf("fetch resource list %s: %w", listURL, err)
	}

	var resList apiResourceList
	if err := json.Unmarshal(listData, &resList); err != nil {
		return nil, fmt.Errorf("parse resource list %s: %w", listURL, err)
	}
	return &resList, nil
}

// splitGroupVersion splits a Kubernetes groupVersion string (e.g. "apps/v1",
// "v1") into its group and version parts. For the core API ("v1"), the group
// is the empty string.
func splitGroupVersion(groupVersion string) (group, version string) {
	if idx := strings.LastIndex(groupVersion, "/"); idx != -1 {
		return groupVersion[:idx], groupVersion[idx+1:]
	}
	return "", groupVersion
}

// kubeVersionPrecedence parses a Kubernetes version string (e.g. "v1",
// "v1beta2", "v2alpha1") into a comparable (tier, major, sub) tuple where
// lower tier = newer (GA=0, beta=1, alpha=2) and higher major/sub = newer
// within a tier.
func kubeVersionPrecedence(v string) (tier, major, sub int) {
	s := strings.TrimPrefix(v, "v")
	if i := strings.Index(s, "alpha"); i != -1 {
		major, _ = strconv.Atoi(s[:i])
		sub, _ = strconv.Atoi(s[i+len("alpha"):])
		return 2, major, sub
	}
	if i := strings.Index(s, "beta"); i != -1 {
		major, _ = strconv.Atoi(s[:i])
		sub, _ = strconv.Atoi(s[i+len("beta"):])
		return 1, major, sub
	}
	major, _ = strconv.Atoi(s)
	return 0, major, 0
}

// kubeVersionIsNewer reports whether Kubernetes version a is newer than b.
// GA > beta > alpha; within a tier, higher version number is newer.
func kubeVersionIsNewer(a, b string) bool {
	ta, ma, sa := kubeVersionPrecedence(a)
	tb, mb, sb := kubeVersionPrecedence(b)
	if ta != tb {
		return ta < tb
	}
	if ma != mb {
		return ma > mb
	}
	return sa > sb
}

// renderURLTemplate renders a FilePath that may contain {{ .version }} using
// the provided version string.
func renderURLTemplate(path helpers.FilePath, version string) (helpers.FilePath, error) {
	tmpl, err := template.New("").Parse(string(path))
	if err != nil {
		return "", fmt.Errorf("parse URL template %q: %w", path, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{"version": version}); err != nil {
		return "", fmt.Errorf("render URL template %q: %w", path, err)
	}
	return helpers.FilePath(buf.String()), nil
}

// matchesFilters returns true when a resource passes the include/exclude lists:
//   - must match at least one Include selector, or Include must be empty
//   - must not match any Exclude selector
func matchesFilters(apiVersion, kind, plural string, includes, excludes []ResourceSelector, pluralLatest map[string]string) bool {
	if len(includes) > 0 {
		matched := false
		for _, sel := range includes {
			if selectorMatches(sel, apiVersion, kind, plural, pluralLatest) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	for _, sel := range excludes {
		if selectorMatches(sel, apiVersion, kind, plural, pluralLatest) {
			return false
		}
	}
	return true
}

// selectorMatches reports whether sel matches a resource identified by
// (apiVersion, kind, plural). Three modes are supported:
//
//   - plural-only (no APIVersion): matches when the plural glob matches and
//     this is the latest Kubernetes version for that resource name.
//   - apiVersion + plural: matches when both glob-match.
//   - apiVersion + kind (or neither): matches when apiVersion glob-matches and
//     at least one kind element glob-matches (empty kind = match everything).
func selectorMatches(sel ResourceSelector, apiVersion, kind, plural string, pluralLatest map[string]string) bool {
	// Mode: plural-only — resolve to latest version.
	if sel.APIVersion == "" && len(sel.Plural) > 0 {
		for _, p := range sel.Plural {
			if globMatch(p, plural) {
				return pluralLatest[plural] == apiVersion
			}
		}
		return false
	}

	// All other modes require the apiVersion to match (if specified).
	if sel.APIVersion != "" && !globMatch(sel.APIVersion, apiVersion) {
		return false
	}

	// Mode: apiVersion + plural.
	if len(sel.Plural) > 0 {
		for _, p := range sel.Plural {
			if globMatch(p, plural) {
				return true
			}
		}
		return false
	}

	// Mode: apiVersion + kind (empty kind = match everything).
	if len(sel.Kind) == 0 {
		return true
	}
	for _, k := range sel.Kind {
		if globMatch(k, kind) {
			return true
		}
	}
	return false
}

// globMatch reports whether pattern matches value.
// The only special character is '*', which matches any sequence of characters.
// All other characters are matched literally.
func globMatch(pattern, value string) bool {
	// Fast paths.
	if pattern == "*" {
		return true
	}
	if !strings.Contains(pattern, "*") {
		return pattern == value
	}

	// Split on '*' and verify each segment appears in order.
	segments := strings.Split(pattern, "*")

	// The first segment must match a prefix of value.
	if segments[0] != "" {
		if !strings.HasPrefix(value, segments[0]) {
			return false
		}
		value = value[len(segments[0]):]
	}

	// The last segment must match a suffix of value.
	last := segments[len(segments)-1]
	if last != "" {
		if !strings.HasSuffix(value, last) {
			return false
		}
		value = value[:len(value)-len(last)]
	}

	// Middle segments must appear in value in order.
	for _, seg := range segments[1 : len(segments)-1] {
		if seg == "" {
			continue
		}
		idx := strings.Index(value, seg)
		if idx == -1 {
			return false
		}
		value = value[idx+len(seg):]
	}

	return true
}
