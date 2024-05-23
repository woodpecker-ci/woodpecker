package woodpecker

import "fmt"

const pathPipelineQueue = "%s/api/pipelines"

// PipelineQueue returns a list of enqueued pipelines.
func (c *client) PipelineQueue() ([]*Feed, error) {
	var out []*Feed
	uri := fmt.Sprintf(pathPipelineQueue, c.addr)
	err := c.get(uri, &out)
	return out, err
}
