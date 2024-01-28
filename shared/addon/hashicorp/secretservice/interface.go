package secretservice

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type Plugin struct {
	Impl model.SecretService
}

func (p *Plugin) Key() string {
	return "configservice"
}

func (p *Plugin) WithImpl(t model.SecretService) plugin.Plugin {
	p.Impl = t
	return p
}

func (p *Plugin) Server(*plugin.MuxBroker) (any, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

func (*Plugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &RPC{client: c}, nil
}
