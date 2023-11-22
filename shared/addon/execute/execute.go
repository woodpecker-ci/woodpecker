package execute

import (
	"github.com/hashicorp/go-plugin"

	"go.woodpecker-ci.org/woodpecker/shared/addon/rpc"
	addonTypes "go.woodpecker-ci.org/woodpecker/shared/addon/types"
)

func Execute[T any](addon addonTypes.Addon[T]) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: rpc.HandshakeConfig,
		Plugins: plugin.PluginSet{
			string(addon.Type()): &rpc.AddonPlugin[T]{Impl: addon},
		},
	})
}
