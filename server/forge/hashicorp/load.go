package hashicorp

import (
	"os/exec"

	"github.com/hashicorp/go-plugin"
)

type Addon[T any] struct {
	Value T
}

func Load[T any](file string, a Plugin[T]) (*Addon[T], error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			a.Key(): a,
		},
		Cmd: exec.Command(file),
	})
	// TODO defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(a.Key())
	if err != nil {
		return nil, err
	}

	extension, _ := raw.(T)
	return &Addon[T]{Value: extension}, nil
}

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "WOODPECKER_PLUGIN",
	MagicCookieValue: "woodpecker-plugin-magic-cookie-value",
}
