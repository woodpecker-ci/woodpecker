package rpc

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"

	"go.woodpecker-ci.org/woodpecker/shared/addon/types"
)

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "woodpecker_plugin_key",
	MagicCookieValue: "woodpecker_plugin_value",
}

type AddonPlugin[T any] struct {
	Impl types.Addon[T]
}

func (a *AddonPlugin[T]) Server(_ *plugin.MuxBroker) (interface{}, error) {
	return &AddonRPCServer[T]{Impl: a.Impl}, nil
}

func (*AddonPlugin[T]) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &AddonRPCClient[T]{client: c}, nil
}
