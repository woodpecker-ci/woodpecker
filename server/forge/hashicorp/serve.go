package hashicorp

import (
	"github.com/hashicorp/go-plugin"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
)

func Serve(impl forge.Forge) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			pluginKey: &Plugin{Impl: impl},
		},
	})
}
