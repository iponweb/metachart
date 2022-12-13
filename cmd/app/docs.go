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

package app

import (
	"bytes"
	"github.com/iponweb/metachart/pkg/chart"
	"sort"
	"strings"
	"text/template"
)

const (
	docsResourcesMdTpl = `
# Resources

A set of resources supported by the chart
{{ range $apiVersion, $resources := .Context }}
## {{ if $apiVersion }}{{ $apiVersion }}{{ else }}Non-Kubernetes resources{{ end }}

| Values file key | Kind | Preprocessor |
| --------------- | ---- | ------------ |
{{ range $r := $resources -}}
| {{ $r.kind }}   | {{ $r.definition.Kind }} | {{ if $r.hasPreprocessor }}[link](templates/preprocess/_{{ $r.kind }}.tpl){{ else }}-{{ end }} |
{{ end }}
{{ end }}
`
)

type templateData struct {
	Context interface{}
}

func (command *GenCommand) GenDocs(c chart.Chart) (err error) {
	t, _ := template.New("resourcedMd").Parse(docsResourcesMdTpl)

	ctx := map[string][]map[string]interface{}{}

	for kind, definition := range c.ResourcesConfig.Resources {
		if !definition.Root {
			continue
		}

		if _, ok := ctx[definition.ApiVersion]; !ok {
			ctx[definition.ApiVersion] = []map[string]interface{}{}
		}

		ctx[definition.ApiVersion] = append(ctx[definition.ApiVersion], map[string]interface{}{
			"kind":            kind,
			"definition":      definition,
			"hasPreprocessor": c.PreprocessorExists(kind),
		})

	}

	for _, resources := range ctx {
		sort.Slice(resources, func(i, j int) bool {
			return resources[i]["kind"].(string) < resources[j]["kind"].(string)
		})
	}

	var b bytes.Buffer
	_ = t.Execute(&b, templateData{
		Context: ctx,
	})

	return c.WriteDocsResourcesMd(strings.TrimSpace(b.String()))
}
