package woodpecker

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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
func (c *client) PipelineMetadata(repoID int64, pipelineNumber int, workflow ...string) ([]byte, error) {
	uri := fmt.Sprintf(pathPipelineMetadata, c.addr, repoID, pipelineNumber)

	if len(workflow) != 0 {
		uri += fmt.Sprintf("?workflow=%s", url.QueryEscape(workflow[0]))
	}

	body, err := c.open(uri, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return io.ReadAll(body)
}
