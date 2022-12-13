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
	"github.com/xanzy/go-gitlab"
	"io"
	"net/http"
	"net/url"
	"os"
)

type FilePath string

const (
	schemaHttp      = "http"
	schemaHttps     = "https"
	schemaFile      = "file"
	schemaGitlabApi = "gitlab-api"

	envVariableGitlabApiToken = "METACHART_GITLAB_API_TOKEN"
)

func (p *FilePath) Read() ([]byte, error) {
	parsed, err := url.Parse(string(*p))
	if err != nil {
		return nil, err
	}

	switch parsed.Scheme {
	case schemaFile:
		return p.ReadFile()
	case schemaHttp, schemaHttps:
		return p.ReadHttp()
	case schemaGitlabApi:
		return p.ReadGitlabApi()
	default:
		return p.ReadFile()
	}
}

func (p *FilePath) ReadFile() ([]byte, error) {
	parsed, err := url.Parse(string(*p))
	if err != nil {
		return nil, err
	}

	return os.ReadFile(parsed.Path)
}

func (p *FilePath) ReadHttp() ([]byte, error) {
	resp, err := http.Get(string(*p))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (p *FilePath) ReadGitlabApi() ([]byte, error) {
	parsed, err := parseGitlabUrl(string(*p))
	if err != nil {
		return nil, err
	}

	//: Explicitly omit reading errors and fail if only token is empty
	token, _ := getGitlabApiToken(parsed.Hostname)
	if token == "" {
		return nil, fmt.Errorf("can not get gitlab api token")
	}

	client, err := gitlab.NewClient(
		token, gitlab.WithBaseURL(fmt.Sprintf("https://%s/api/v4", parsed.Hostname)))
	if err != nil {
		return nil, err
	}

	file, _, err := client.RepositoryFiles.GetRawFile(
		parsed.ProjectPath, parsed.FilePath,
		&gitlab.GetRawFileOptions{
			Ref: &parsed.Revision,
		})
	if err != nil {
		return nil, err
	}

	return file, nil
}
