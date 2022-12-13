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
	configResourcesYamlPath          = "config/resources.yaml"
	configSchemaYamlPath             = "config/schema.yaml"
	configValuesSchemaCustomJsonPath = "config/values.schema.custom.json"
	templatesCustomTplPath           = "templates/_custom.tpl"
	templatesMetachartTplPath        = "templates/_metachart.tpl"
	templatesResourcesYamlPath       = "templates/resources.yaml"
)

type ConversionRule struct {
	Source     *string            `json:"source"`
	Target     string             `json:"target"`
	Properties *map[string]string `json:"properties,omitempty"`
	Allowed    *[]string          `json:"allowed,omitempty"`
	Disallowed *[]string          `json:"disallowed,omitempty"`
	Required   *[]string          `json:"required,omitempty"`
	Related    map[string]string  `json:"related"`
}

type SchemaConfig struct {
	Definitions []helpers.FilePath `json:"definitions"`
	Rules       []ConversionRule   `json:"rules"`
}

type ResourceDefinition struct {
	Template      bool   `json:"template"`
	ApiVersion    string `json:"apiVersion"`
	Kind          string `json:"kind"`
	JsonSchemaRef string `json:"jsonSchemaRef"`
	Root          bool   `json:"root"`
	Defaults      bool   `json:"defaults"`
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

type ResourcesConfig struct {
	Resources map[string]ResourceDefinition `json:"resources"`
}

type Chart struct {
	SchemaConfig    SchemaConfig
	ResourcesConfig ResourcesConfig

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
		chart.SchemaConfig.Definitions,
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
	var (
		schemaConfig    = SchemaConfig{}
		resourcesConfig = ResourcesConfig{}
		err             error
	)
	err = helpers.ReadYamlFile(filepath.Join(root, configSchemaYamlPath), &schemaConfig)
	if err != nil {
		return nil, err
	}

	err = helpers.ReadYamlFile(filepath.Join(root, configResourcesYamlPath), &resourcesConfig)
	if err != nil {
		return nil, err
	}

	return &Chart{
		SchemaConfig:    schemaConfig,
		ResourcesConfig: resourcesConfig,
		root:            root,
	}, nil
}

func NewChartEmpty(root string) *Chart {
	return &Chart{
		root: root,
	}
}
