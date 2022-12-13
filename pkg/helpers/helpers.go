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

package helpers

import (
	"os"
	"sigs.k8s.io/yaml"
)

func SliceIndex[T comparable](s []T, e T) int {
	for i, element := range s {
		if element == e {
			return i
		}
	}
	return -1
}

func SliceContains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func SliceUnique[T comparable](s []T) []T {
	keys := make(map[T]bool)
	var list []T
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func ReadYamlFile(path string, o interface{}) error {
	yamlFile, err := os.ReadFile(path)
	if err == nil {
		return yaml.Unmarshal(yamlFile, o)
	}
	return err
}
