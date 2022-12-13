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

import _ "embed"

var (
	//go:embed resources/init/.helmignore
	helmignoreBody string

	//go:embed resources/init/Chart.yaml
	chartYamlBody string

	//go:embed resources/init/values.yaml
	valuesYamlBody string

	//go:embed resources/init/config/resources.yaml
	configResourcesYamlBody string

	//go:embed resources/init/config/schema.yaml
	configSchemaYamlBody string

	//go:embed resources/init/config/values.schema.custom.json
	configValuesSchemaCustomJsonBody string

	//go:embed resources/init/templates/_custom.tpl
	templatesCustomTplBody string

	//go:embed resources/init/templates/_metachart.tpl
	templatesMetachartTplBody string

	//go:embed resources/init/templates/resources.yaml
	templatesResourcesYamlBody string

	initData = map[string]string{
		helmignorePath:                   helmignoreBody,
		chartYamlPath:                    chartYamlBody,
		valuesYamlPath:                   valuesYamlBody,
		configResourcesYamlPath:          configResourcesYamlBody,
		configSchemaYamlPath:             configSchemaYamlBody,
		configValuesSchemaCustomJsonPath: configValuesSchemaCustomJsonBody,
		templatesCustomTplPath:           templatesCustomTplBody,
		templatesMetachartTplPath:        templatesMetachartTplBody,
		templatesResourcesYamlPath:       templatesResourcesYamlBody,
	}

	genData = map[string]string{
		templatesMetachartTplPath:  templatesMetachartTplBody,
		templatesResourcesYamlPath: templatesResourcesYamlBody,
	}
)
