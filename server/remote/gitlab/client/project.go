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
	"encoding/json"
	"strings"
)

const (
	projectsUrl       = "/projects"
	commitStatusUrl   = "/projects/:id/statuses/:sha"
)



// SetStatus report ci status of a specific commit
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
