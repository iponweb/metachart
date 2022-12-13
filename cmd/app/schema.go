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
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/barkimedes/go-deepcopy"
	"github.com/iponweb/metachart/pkg/chart"
	"github.com/iponweb/metachart/pkg/helpers"
	"strings"
)

const (
	FqdnName             = "^[a-z][0-9a-z]*(-[0-9a-z]+)*$"
	checksumsRef         = "metachart.interface.checksums"
	checksumEntryListRef = "metachart.interface.checksumEntryList"
)

//go:embed resources/values.schema.base.json
var baseSchema string

func GenRootKeyProperty(ref string) map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"patternProperties": map[string]interface{}{
			FqdnName: GenReferenceDefinition(ref),
		},
		"additionalProperties": false,
	}
}

func GenEmptyObjectDefinition() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func GenReferenceDefinition(ref string) map[string]interface{} {
	return map[string]interface{}{
		"$ref": "#/definitions/" + ref,
	}
}

func GenDisabledDefinition() map[string]interface{} {
	return map[string]interface{}{
		"type":        "boolean",
		"description": "Disable all resources of this kind",
	}
}

func GetDefinitionProperties(definition map[string]interface{}) map[string]interface{} {
	if value, ok := definition["properties"]; ok {
		if properties, ok := value.(map[string]interface{}); ok {
			return properties
		}
	}
	return nil
}

func GetPropertiesEntry(properties map[string]interface{}, name string) map[string]interface{} {
	if valueRaw, ok := properties[name]; ok {
		if value, ok := valueRaw.(map[string]interface{}); ok {
			return value
		}
	}
	return nil
}

func GetDefinitionRequired(definition map[string]interface{}) *[]string {
	var required []string
	if value, ok := definition["required"]; ok {
		if s, ok := value.([]interface{}); ok {
			for _, value := range s {
				if entry, ok := value.(string); ok {
					required = append(required, entry)
				} else {
					return nil
				}
			}
			return &required
		} else if s, ok := value.([]string); ok {
			return &s
		}
	}
	return nil
}

func CollectDefinitionRefs(data interface{}) (result []string) {
	if parsed, ok := data.(map[string]interface{}); ok {
		for key, valueRaw := range parsed {
			if value, ok := valueRaw.(string); ok {
				if key == "$ref" {
					if strings.HasPrefix(value, "#/definitions/") {
						result = append(result, strings.TrimPrefix(value, "#/definitions/"))
					}
				}
			} else if _, ok := valueRaw.(map[string]interface{}); ok {
				result = append(result, CollectDefinitionRefs(valueRaw)...)
			} else if _, ok := valueRaw.([]interface{}); ok {
				result = append(result, CollectDefinitionRefs(valueRaw)...)
			}
		}
	} else if parsed, ok := data.([]interface{}); ok {
		for _, item := range parsed {
			result = append(result, CollectDefinitionRefs(item)...)
		}
	}
	return
}

func CollectUsedDefinitions(data map[string]interface{}, allDefinitions map[string]interface{}) []string {
	result := helpers.SliceUnique(CollectDefinitionRefs(data))
	toCheck := result

	for len(toCheck) > 0 {
		var toCheckNew []string

		for _, item := range toCheck {
			for _, ref := range CollectDefinitionRefs(allDefinitions[item]) {
				if !helpers.SliceContains(result, ref) {
					toCheckNew = append(toCheckNew, ref)
					result = append(result, ref)
				}
			}
		}

		toCheck = helpers.SliceUnique(toCheckNew)
	}

	return result
}

func CleanupDescriptions(schema chart.JsonSchema) chart.JsonSchema {
	for _, definitionRaw := range schema.Definitions {
		if definition, ok := definitionRaw.(map[string]interface{}); ok {
			if _, ok := definition["description"]; ok {
				delete(definition, "description")
			}

			if propertiesRaw, ok := definition["properties"]; ok {
				if properties, ok := propertiesRaw.(map[string]interface{}); ok {
					for _, propertyDefinitionRaw := range properties {
						if propertyDefinition, ok := propertyDefinitionRaw.(map[string]interface{}); ok {
							if _, ok := propertyDefinition["description"]; ok {
								delete(propertyDefinition, "description")
							}
						}
					}
				}
			}
		}
	}

	return schema
}

func (command *GenCommand) GenSchema(c chart.Chart) error {
	var schema chart.JsonSchema

	err := json.Unmarshal([]byte(baseSchema), &schema)
	if err != nil {
		return err
	}

	allDefinitions, err := c.ReadDefinitions()
	if err != nil {
		return err
	}
	for _, s := range *allDefinitions {
		for ref, definition := range s.Definitions {
			schema.Definitions[ref] = definition
		}
	}

	for _, rule := range c.SchemaConfig.Rules {
		var definition map[string]interface{}

		//: Get property definition by source or build empty
		if rule.Source == nil {
			definition = GenEmptyObjectDefinition()
		} else if valueRaw, ok := schema.Definitions[*rule.Source]; ok {
			valueRaw = deepcopy.MustAnything(valueRaw)
			if value, ok := valueRaw.(map[string]interface{}); ok {
				definition = value
			} else {
				return fmt.Errorf("can not find sourceDefinition '%s'", *rule.Source)
			}
		} else {
			return fmt.Errorf("can not find sourceDefinition '%s'", *rule.Source)
		}

		//: Disallowed
		if rule.Disallowed != nil && len(*rule.Disallowed) > 0 {
			properties := GetDefinitionProperties(definition)
			if properties == nil {
				properties = map[string]interface{}{}
			}
			newProperties := map[string]interface{}{}
			for key, value := range properties {
				if !helpers.SliceContains(*rule.Disallowed, key) {
					newProperties[key] = value
				}
			}
			definition["properties"] = newProperties

			required := GetDefinitionRequired(definition)
			if required != nil {
				var newRequired []string
				for _, value := range *required {
					if !helpers.SliceContains(*rule.Disallowed, value) {
						newRequired = append(newRequired, value)
					}
				}
				definition["required"] = newRequired
			}
		}

		//: Allowed
		if rule.Allowed != nil && len(*rule.Allowed) > 0 {
			properties := GetDefinitionProperties(definition)
			if properties == nil {
				properties = map[string]interface{}{}
			}
			newProperties := map[string]interface{}{}
			for key, value := range properties {
				if helpers.SliceContains(*rule.Allowed, key) {
					newProperties[key] = value
				}
			}
			definition["properties"] = newProperties

			required := GetDefinitionRequired(definition)
			if required != nil {
				var newRequired []string
				for _, value := range *required {
					if helpers.SliceContains(*rule.Allowed, value) {
						newRequired = append(newRequired, value)
					}
				}
				definition["required"] = newRequired
			}
		}

		//: Properties
		if rule.Properties != nil {
			properties := GetDefinitionProperties(definition)
			if properties == nil {
				properties = map[string]interface{}{}
			}
			newProperties := properties

			for key, value := range *rule.Properties {
				newProperties[key] = GenReferenceDefinition(value)
			}
			definition["properties"] = newProperties
		}

		//: Required
		if rule.Required != nil {
			required := GetDefinitionRequired(definition)
			if required == nil {
				required = &[]string{}
			}
			definition["required"] = append(*required, *rule.Required...)
		}
		//: Cleanup empty and nil value
		if _, ok := definition["required"]; ok {
			required := GetDefinitionRequired(definition)
			if required == nil || len(*required) == 0 {
				delete(definition, "required")
			}
		}

		//: Related
		if len(rule.Related) > 0 {
			relatedProperties := map[string]interface{}{}

			for kind, ref := range rule.Related {
				relatedProperties[kind] = GenRootKeyProperty(ref)
			}

			properties := GetDefinitionProperties(definition)
			if properties == nil {
				properties = map[string]interface{}{}
			}
			related := GenEmptyObjectDefinition()
			related["properties"] = relatedProperties
			properties["related"] = related
			definition["properties"] = properties
		}

		schema.Definitions[rule.Target] = definition
	}

	//: Kinds
	for kind, definition := range c.ResourcesConfig.Resources {
		if definition.Root {
			schema.Properties[kind] = GenRootKeyProperty(definition.JsonSchemaRef)
		}
	}

	//: Settings
	settings := GetPropertiesEntry(schema.Properties, "settings")
	if settings == nil {
		settings = map[string]interface{}{}
		schema.Properties["settings"] = settings
	}

	settingsProperties := GetDefinitionProperties(settings)
	if settingsProperties == nil {
		settingsProperties = map[string]interface{}{}
		settings["properties"] = settingsProperties
	}

	for kind, definition := range c.ResourcesConfig.Resources {
		if !(definition.Root || definition.Defaults) {
			continue
		}

		kindSettings := GetPropertiesEntry(settingsProperties, kind)
		if kindSettings == nil {
			kindSettings = map[string]interface{}{
				"type": "object",
			}
			settingsProperties[kind] = kindSettings
		}

		kindSettingsProperties := GetDefinitionProperties(kindSettings)
		if kindSettingsProperties == nil {
			kindSettingsProperties = map[string]interface{}{}
			kindSettings["properties"] = kindSettingsProperties
		}

		if definition.Root {
			kindSettingsProperties["disabled"] = GenDisabledDefinition()
		}
		if definition.Defaults {
			kindSettingsProperties["defaults"] = GenReferenceDefinition(definition.JsonSchemaRef)
		}
	}

	//: Checksums
	checksums := GenEmptyObjectDefinition()
	schema.Definitions[checksumsRef] = checksums

	checksumsProperties := map[string]interface{}{}
	checksums["properties"] = checksumsProperties

	for kind, definition := range c.ResourcesConfig.Resources {
		if !definition.Root {
			continue
		}
		checksumsProperties[kind] = GenReferenceDefinition(checksumEntryListRef)
	}

	//: Cleanup unused definitions
	usedDefinitions := CollectUsedDefinitions(schema.Properties, schema.Definitions)
	for key, _ := range schema.Definitions {
		if !helpers.SliceContains(usedDefinitions, key) {
			delete(schema.Definitions, key)
		}
	}

	err = c.WriteSchemaFull(schema)
	if err != nil {
		return err
	}
	return c.WriteSchema(CleanupDescriptions(schema))
}
