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
	"fmt"
	"net/url"
	"os"
	"strings"
)

type GitlabFileUrl struct {
	Hostname    string
	ProjectPath string
	FilePath    string
	Revision    string
}

func (u *GitlabFileUrl) String() string {
	return fmt.Sprintf("%s %s %s %s", u.Hostname, u.ProjectPath, u.Revision, u.FilePath)
}

func parseGitlabUrl(u string) (*GitlabFileUrl, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	elements := strings.Split(parsed.Path, "/")
	idx := SliceIndex(elements, "-")
	if idx == -1 {
		return nil, fmt.Errorf("can not parse gitlab url")
	}

	if len(elements) < idx+3 || idx < 2 {
		return nil, fmt.Errorf("can not parse gitlab url")
	}

	return &GitlabFileUrl{
		Hostname:    parsed.Hostname(),
		ProjectPath: strings.Join(elements[1:idx], "/"),
		Revision:    elements[idx+2],
		FilePath:    strings.Join(elements[idx+3:], "/"),
	}, nil
}

func getGitlabApiToken(hostname string) (string, error) {
	return os.Getenv(envVariableGitlabApiToken), nil
}
