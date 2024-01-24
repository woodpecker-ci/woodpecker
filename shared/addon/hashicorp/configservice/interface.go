package configservice

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"

	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/config"
)

type ExtensionPlugin struct {
	Impl config.Extension
}

func (p *ExtensionPlugin) Key() string {
	return "configservice"
}

func (p *ExtensionPlugin) WithImpl(t config.Extension) plugin.Plugin {
	p.Impl = t
	return p
}

func (p *ExtensionPlugin) Server(*plugin.MuxBroker) (any, error) {
	return &ExtensionRPCServer{Impl: p.Impl}, nil
}

func (*ExtensionPlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &ExtensionRPC{client: c}, nil
}
