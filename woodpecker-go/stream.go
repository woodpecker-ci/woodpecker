package woodpeckergo

import (
	"context"
	"fmt"

	"github.com/r3labs/sse/v2"
)

func (c *Client) StreamPipelineStepLog(ctx context.Context, repoID, pipelineNumber, stepID int64) error {
	logs := make(chan *sse.Event)

	client := sse.NewClient(fmt.Sprintf("%s/api/stream/logs/%d/%d/%d", c.uri, repoID, pipelineNumber, stepID))
	client.Connection.Transport = c.transport.Transport

	client.SubscribeChan("logs", logs)
	defer client.Close()
	<-ctx.Done()
	client.Unsubscribe(logs)

	return nil
}
