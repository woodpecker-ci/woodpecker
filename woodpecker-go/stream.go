package woodpeckergo

import (
	"fmt"

	"github.com/r3labs/sse/v2"
)

func (c *Client) StreamPipelineStepLog(repoID, pipelineNumber, stepID int64) error {
	events := make(chan *sse.Event)

	client := sse.NewClient(fmt.Sprintf("%s/api/stream/logs/%d/%d/%d", c.uri, repoID, pipelineNumber, stepID))
	client.Connection.Transport = c.transport.Transport

	return client.SubscribeChan("logs", events)
}
