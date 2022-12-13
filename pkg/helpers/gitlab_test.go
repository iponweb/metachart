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
	"testing"
)

func Test_parseGitlabUrl(t *testing.T) {
	t.Run("ok", func(in *testing.T) {
		result, err := parseGitlabUrl("https://gitlab.example.net/path/to/project/-/blob/master/README.rst")

		expected := GitlabFileUrl{
			Hostname:    "gitlab.example.net",
			ProjectPath: "path/to/project",
			Revision:    "master",
			FilePath:    "README.rst",
		}

		if err != nil {
			in.Errorf("unexpected err")
		} else if result == nil {
			in.Errorf("unexpected nil returned")
		} else if *result != expected {
			in.Errorf("expected: `%s` got: `%s`", expected.String(), result.String())
		}
	})

	t.Run("ok", func(in *testing.T) {
		result, err := parseGitlabUrl("https://gitlab.example.net/path/to/project/-/raw/master/README.rst")

		expected := GitlabFileUrl{
			Hostname:    "gitlab.example.net",
			ProjectPath: "path/to/project",
			Revision:    "master",
			FilePath:    "README.rst",
		}

		if err != nil {
			in.Errorf("unexpected err")
		} else if result == nil {
			in.Errorf("unexpected nil returned")
		} else if *result != expected {
			in.Errorf("expected: `%s` got: `%s`", expected.String(), result.String())
		}
	})

	t.Run("broken", func(in *testing.T) {
		_, err := parseGitlabUrl("https://gitlab.example.net/path/to/project/-")

		if err == nil {
			in.Errorf("expected error")
		}
	})

	t.Run("broken", func(in *testing.T) {
		_, err := parseGitlabUrl("https://gitlab.example.net/-/raw/README.rst")

		if err == nil {
			in.Errorf("expected error")
		}
	})

}
