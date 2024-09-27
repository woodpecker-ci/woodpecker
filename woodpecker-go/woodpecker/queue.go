package woodpecker

import "fmt"

const pathQueue = "%s/api/queue"

// QueueInfo returns queue info.
func (c *client) QueueInfo() (*Info, error) {
	out := new(Info)
	uri := fmt.Sprintf(pathQueue+"/info", c.addr)
	err := c.get(uri, out)
	return out, err
}
