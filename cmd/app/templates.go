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
	"github.com/iponweb/metachart/pkg/chart"
	"sigs.k8s.io/yaml"
	"strings"
)

const resourcesYamlTpl = `
{{- /* Resources definition */}}
{{- define "metachart.settings" }}
SETTINGS
{{- end }}
`

type KindSettings struct {
	ApiVersion    string `json:"apiVersion"`
	KindCamelCase string `json:"kindCamelCase"`
	Preprocess    bool   `json:"preprocess"`
}

type Settings map[string]KindSettings

func (command *GenCommand) GenTemplates(c chart.Chart) (err error) {
	settings := Settings{}

	for kind, config := range c.ResourcesConfig.Resources {
		if !config.Template {
			continue
		}

		settings[kind] = KindSettings{
			ApiVersion:    config.ApiVersion,
			KindCamelCase: config.Kind,
			Preprocess:    c.PreprocessorExists(kind),
		}
	}

	renderedSettings, err := yaml.Marshal(&settings)
	if err != nil {
		return err
	}

	rendered := resourcesYamlTpl
	rendered = strings.Replace(rendered, "SETTINGS", strings.TrimSpace(string(renderedSettings)), -1)
	rendered = strings.TrimSpace(rendered)

	err = c.WriteTemplate("_settings.tpl", rendered)
	if err != nil {
		return err
	}

	return nil
}
