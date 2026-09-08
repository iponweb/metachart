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
	"encoding/json"
	"github.com/iponweb/metachart/pkg/helpers"
	"os"
	"path"
	"path/filepath"
)

const (
	//: Chart managed layout

	//: Generated files
	valuesSchemaJsonPath              = "values.schema.json"
	valuesSchemaFullJsonPath          = "values.schema.full.json"
	templatesGeneratedDir             = "templates/generated"
	templatesGeneratedSettingsTplPath = "templates/generated/_settings.tpl"
	templatesPreprocessDir            = "templates/preprocess"
	docsResourcesMdPath               = "docs/resources.md"

	//: Static content files
	helmignorePath                   = ".helmignore"
	chartYamlPath                    = "Chart.yaml"
	valuesYamlPath                   = "values.yaml"
	configYamlPath                   = "config/config.yaml"
	configValuesSchemaCustomJsonPath = "config/values.schema.custom.json"
	templatesCustomTplPath           = "templates/_custom.tpl"
	templatesMetachartTplPath        = "templates/_metachart.tpl"
	templatesResourcesYamlPath       = "templates/resources.yaml"
)

// ConversionRule describes how to derive a metachart schema definition from a
// source JSON Schema definition. All filter fields (Allowed, Disallowed,
// Required, Properties, Related) are optional.
type ConversionRule struct {
	Source     *string            `json:"source"`
	Target     string             `json:"target"`
	Properties *map[string]string `json:"properties,omitempty"`
	Allowed    *[]string          `json:"allowed,omitempty"`
	Disallowed *[]string          `json:"disallowed,omitempty"`
	Required   *[]string          `json:"required,omitempty"`
	Related    map[string]string  `json:"related"`
}

// SchemaConfig holds the JSON Schema source definitions and the conversion
// rules that transform them into metachart-specific definitions.
type SchemaConfig struct {
	Definitions []helpers.FilePath `json:"definitions"`
	Rules       []ConversionRule   `json:"rules"`
}

// ResourceDefinition describes a single Kubernetes resource kind that the
// chart manages. Template, Root and Defaults default to true when omitted.
// Global defaults to false: when set, `global.<kind>` is merged into the root
// key `<kind>` before rendering by `metachart.applyGlobal`, the same way
// `global.settings` is always merged into `settings`. Requires Root.
type ResourceDefinition struct {
	Template      bool   `json:"template"`
	ApiVersion    string `json:"apiVersion"`
	Kind          string `json:"kind"`
	JsonSchemaRef string `json:"jsonSchemaRef"`
	Root          bool   `json:"root"`
	Defaults      bool   `json:"defaults"`
	Global        bool   `json:"global"`
}

func (c *ResourceDefinition) UnmarshalJSON(data []byte) error {
	type Alias ResourceDefinition
	type Aux struct {
		Template *bool `json:"template"`
		Root     *bool `json:"root"`
		Defaults *bool `json:"defaults"`
		*Alias
	}
	aux := &Aux{Alias: (*Alias)(c)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Template == nil {
		c.Template = true
	} else {
		c.Template = *aux.Template
	}

	if aux.Root == nil {
		c.Root = true
	} else {
		c.Root = *aux.Root
	}

	if aux.Defaults == nil {
		c.Defaults = true
	} else {
		c.Defaults = *aux.Defaults
	}

	return nil
}

// ResourceSelector matches a Kubernetes resource. Three modes are supported,
// determined by which fields are set. Kind and Plural are mutually exclusive.
//
//   - apiVersion + kind:   match by API version and kind name(s)
//   - apiVersion + plural: match by API version and plural resource name(s)
//   - plural (no apiVersion): match by plural name, picking the latest
//     Kubernetes version automatically (GA > beta > alpha, higher wins)
//
// All string fields support '*' as a wildcard. Omitting both kind and plural
// with only apiVersion set matches every resource in that API version.
//
// The optional Related map adds entries to the ConversionRule's related field
// for every resource matched by this selector. Each key is the relationship
// name (plural) and the value is a selector that resolves to the related
// resource's schema reference. Resolution is performed after all autodiscover
// sources have been processed, so cross-source references work.
//
// Examples:
//
//	{apiVersion: "apps/v1", kind: ["Deployment", "DaemonSet"]}
//	{apiVersion: "apps/v1", plural: ["deployments"]}
//	{plural: ["deployments", "daemonsets"]}   — latest version, auto-detected
//	{apiVersion: "autoscaling/v2", kind: ["HorizontalPodAutoscaler"],
//	  related: {deployments: {plural: ["deployments"]}}}
type ResourceSelector struct {
	APIVersion string   `json:"apiVersion,omitempty"`
	Kind       []string `json:"kind,omitempty"`
	Plural     []string `json:"plural,omitempty"`
	// Disallowed fields to append to the autodiscovered ConversionRule's
	// disallowed list on top of the defaults (status, kind, apiVersion).
	Disallowed []string `json:"disallowed,omitempty"`
	// Properties to merge into the autodiscovered ConversionRule on top of the
	// defaults (enabled, metadata). Later keys override earlier ones.
	Properties map[string]string `json:"properties,omitempty"`
	// Related selectors whose matched plural names become the relationship keys
	// in the ConversionRule. Each selector resolves to one resource; the plural
	// name of that resource is used as the key.
	Related []ResourceSelector `json:"related,omitempty"`
}

// AutodiscoverSource configures automatic resource discovery from a Kubernetes
// API discovery dump and a matching JSON Schema definitions file.
//
// Include and Exclude are lists of ResourceSelectors. A resource is included
// when it matches at least one Include entry (or Include is empty) and does
// not match any Exclude entry.
//
// The optional Version field is substituted as {{ .version }} inside the
// Discovery and Definitions URL strings before they are fetched, making it
// easy to pin or bump a version in one place:
//
//	version: "v1.5.1"
//	discovery: "https://…/{{ .version }}/discovery/apis.json"
//	definitions: "https://…/{{ .version }}/json-schema/source/_definitions.json"
type AutodiscoverSource struct {
	// Version is an optional version string available as {{ .version }} in
	// the Discovery and Definitions URL templates.
	Version string `json:"version,omitempty"`
	// Discovery is the URL (or URL template) of the apis.json discovery file.
	Discovery helpers.FilePath `json:"discovery"`
	// Definitions is the URL (or URL template) of the _definitions.json file.
	Definitions helpers.FilePath `json:"definitions"`
	// Include lists selectors for resources to include.
	// An empty list includes all discovered resources.
	Include []ResourceSelector `json:"include,omitempty"`
	// Exclude lists selectors for resources to exclude after include filtering.
	Exclude []ResourceSelector `json:"exclude,omitempty"`
	// DefinitionsLast moves this source's definitions URL to the end of the
	// merged definitions list, so its type definitions take precedence over
	// all other autodiscover sources. Use this for Kubernetes core definitions
	// to prevent CRD bundles (which often embed incomplete copies of core
	// Kubernetes types) from shadowing the authoritative definitions.
	DefinitionsLast bool `json:"definitionsLast,omitempty"`
}

// MetachartConfigSpec is the spec section of a MetachartConfig object.
type MetachartConfigSpec struct {
	Schema       SchemaConfig                  `json:"schema"`
	Autodiscover []AutodiscoverSource          `json:"autodiscover,omitempty"`
	Resources    map[string]ResourceDefinition `json:"resources"`
}

// ObjectMeta holds identifying metadata for a MetachartConfig object.
type ObjectMeta struct {
	Name string `json:"name"`
}

// MetachartConfig is the top-level configuration object. It follows the
// Kubernetes object convention (apiVersion / kind / metadata / spec) so that
// editors with schema support can validate it in the same way they validate
// other manifests.
//
// Example:
//
//	apiVersion: metachart.iponweb.net/v1alpha1
//	kind: MetachartConfig
//	metadata:
//	  name: my-chart
//	spec:
//	  schema:
//	    definitions: [...]
//	    rules: [...]
//	  resources:
//	    deployments:
//	      apiVersion: apps/v1
//	      kind: Deployment
//	      jsonSchemaRef: metachart.api.io.k8s.api.apps.v1.Deployment
type MetachartConfig struct {
	APIVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	Metadata   ObjectMeta          `json:"metadata"`
	Spec       MetachartConfigSpec `json:"spec"`
}

type Chart struct {
	Config MetachartConfig

	root string
}

func (chart *Chart) CleanupTemplates() error {
	path := filepath.Join(chart.root, templatesGeneratedDir)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}

	return os.MkdirAll(path, os.ModePerm)
}

func (chart *Chart) PreprocessorExists(kind string) bool {
	path := filepath.Join(chart.root, templatesPreprocessDir, "_"+kind+".tpl")
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (chart *Chart) WriteTemplate(name, body string) error {
	return chart.writeFile(filepath.Join(templatesGeneratedDir, name), []byte(body))
}

func (chart *Chart) WriteSchema(schema JsonSchema) error {
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}
	return chart.writeFile(valuesSchemaJsonPath, data)
}

func (chart *Chart) WriteSchemaFull(schema JsonSchema) error {
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}
	return chart.writeFile(valuesSchemaFullJsonPath, data)
}

func (chart *Chart) WriteDocsResourcesMd(body string) error {
	return chart.writeFile(docsResourcesMdPath, []byte(body))
}

func (chart *Chart) WriteSettings(body string) error {
	return chart.writeFile(templatesGeneratedSettingsTplPath, []byte(body))
}

func (chart *Chart) writeFile(p string, b []byte) error {
	absPath := filepath.Join(chart.root, p)

	err := os.MkdirAll(path.Dir(absPath), os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(absPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(b)
	return err
}

func (chart *Chart) ReadDefinitions() (*[]JsonSchema, error) {
	var result []JsonSchema

	paths := append(
		chart.Config.Spec.Schema.Definitions,
		helpers.FilePath(filepath.Join(chart.root, configValuesSchemaCustomJsonPath)))

	for _, definitionsPath := range paths {
		var entry JsonSchema

		jsonFile, err := definitionsPath.Read()
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(jsonFile, &entry); err != nil {
			return nil, err
		}

		result = append(result, entry)
	}

	return &result, nil
}

func (chart *Chart) WriteInit() error {
	for p, b := range initData {
		err := chart.writeFile(p, []byte(b))
		if err != nil {
			return err
		}
	}
	return nil
}

func (chart *Chart) WriteGen() error {
	for p, b := range genData {
		err := chart.writeFile(p, []byte(b))
		if err != nil {
			return err
		}
	}
	return nil
}

func (chart *Chart) IsEmpty() (bool, error) {
	var files []string
	err := filepath.Walk(chart.root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	return len(files) == 1, err
}

func NewChart(root string) (*Chart, error) {
	config := MetachartConfig{}

	err := helpers.ReadYamlFile(filepath.Join(root, configYamlPath), &config)
	if err != nil {
		return nil, err
	}

	return &Chart{
		Config: config,
		root:   root,
	}, nil
}

func NewChartEmpty(root string) *Chart {
	return &Chart{
		root: root,
	}
}
