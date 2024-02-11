package hashicorp

import (
	"github.com/hashicorp/go-plugin"
)

func Serve[T any](addon Plugin[T], impl T) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			addon.Key(): addon.WithImpl(impl),
		},
	})
}
