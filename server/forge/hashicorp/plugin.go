package hashicorp

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
)

const pluginKey = "forge"

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "WOODPECKER_FORGE_ADDON_PLUGIN",
	MagicCookieValue: "woodpecker-plugin-magic-cookie-value",
}

type Plugin struct {
	Impl forge.Forge
}

func (p *Plugin) Server(*plugin.MuxBroker) (any, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

func (*Plugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &RPC{client: c}, nil
}
