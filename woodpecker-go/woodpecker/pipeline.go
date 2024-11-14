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
