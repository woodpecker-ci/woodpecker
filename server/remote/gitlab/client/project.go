// Copyright 2018 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
)

const (
	projectsUrl       = "/projects"
	projectUrl        = "/projects/:id"
	repoUrlRawFileRef = "/projects/:id/repository/files/:filepath"
	commitStatusUrl   = "/projects/:id/statuses/:sha"
)

// Get a list of all projects owned by the authenticated user.
func (g *Client) AllProjects(hide_archives bool) ([]*Project, error) {
	var per_page = 100
	var projects []*Project

	for i := 1; true; i++ {
		contents, err := g.Projects(i, per_page, hide_archives)
		if err != nil {
			return projects, err
		}

		for _, value := range contents {
			projects = append(projects, value)
		}

		if len(projects) == 0 {
			break
		}

		if len(projects)/i < per_page {
			break
		}
	}

	return projects, nil
}

// Get a list of projects owned by the authenticated user.
func (c *Client) Projects(page int, per_page int, hide_archives bool) ([]*Project, error) {
	projectsOptions := QMap{
		"page":       strconv.Itoa(page),
		"per_page":   strconv.Itoa(per_page),
		"membership": "true",
	}

	if hide_archives {
		projectsOptions["archived"] = "false"
	}

	url, opaque := c.ResourceUrl(projectsUrl, nil, projectsOptions)

	var projects []*Project

	contents, err := c.Do("GET", url, opaque, nil)
	if err == nil {
		err = json.Unmarshal(contents, &projects)
	}

	return projects, err
}

// Get a project by id
func (c *Client) Project(id string) (*Project, error) {
	url, opaque := c.ResourceUrl(projectUrl, QMap{":id": id}, nil)

	var project *Project

	contents, err := c.Do("GET", url, opaque, nil)
	if err == nil {
		err = json.Unmarshal(contents, &project)
	}

	return project, err
}

func (c *Client) RepoRawFileRef(id, ref, filepath string) ([]byte, error) {
	var fileRef FileRef
	url, opaque := c.ResourceUrl(
		repoUrlRawFileRef,
		QMap{
			":id":       id,
			":filepath": filepath,
		},
		QMap{
			"ref": ref,
		},
	)

	contents, err := c.Do("GET", url, opaque, nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &fileRef)
	if err != nil {
		return nil, err
	}

	fileRawContent, err := base64.StdEncoding.DecodeString(fileRef.Content)
	return fileRawContent, err
}

//
func (c *Client) SetStatus(id, sha, state, desc, ref, link string) error {
	url, opaque := c.ResourceUrl(
		commitStatusUrl,
		QMap{
			":id":  id,
			":sha": sha,
		},
		QMap{
			"state":       state,
			"ref":         ref,
			"target_url":  link,
			"description": desc,
			"context":     "ci/drone",
		},
	)

	_, err := c.Do("POST", url, opaque, nil)
	return err
}

// Get a list of projects by query owned by the authenticated user.
func (c *Client) SearchProjectId(namespace string, name string) (id int, err error) {

	url, opaque := c.ResourceUrl(projectsUrl, nil, QMap{
		"query":      strings.ToLower(name),
		"membership": "true",
	})

	var projects []*Project

	contents, err := c.Do("GET", url, opaque, nil)
	if err == nil {
		err = json.Unmarshal(contents, &projects)
	} else {
		return id, err
	}

	for _, project := range projects {
		if project.Namespace.Name == namespace && strings.ToLower(project.Name) == strings.ToLower(name) {
			id = project.Id
		}
	}

	return id, err
}
