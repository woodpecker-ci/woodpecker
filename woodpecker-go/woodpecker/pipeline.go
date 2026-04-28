// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package woodpecker

import (
	"fmt"
	"io"
	"net/http"
)

const (
	pathPipelineQueue    = "%s/api/pipelines"
	pathPipelineMetadata = "%s/api/repos/%d/pipelines/%d/metadata"
)

// PipelineQueue returns a list of enqueued pipelines.
func (c *client) PipelineQueue() ([]*Feed, error) {
	var out []*Feed
	uri := fmt.Sprintf(pathPipelineQueue, c.addr)
	err := c.get(uri, &out)
	return out, err
}

// PipelineMetadata returns metadata for a pipeline, workflow name is optional.
func (c *client) PipelineMetadata(repoID int64, pipelineNumber int) ([]byte, error) {
	uri := fmt.Sprintf(pathPipelineMetadata, c.addr, repoID, pipelineNumber)

	body, err := c.open(uri, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return io.ReadAll(body)
}
