package hashicorp

import (
	"os/exec"

	"github.com/hashicorp/go-plugin"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
)

const pluginKey = "forge"

func Load(file string) (forge.Forge, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			pluginKey: &Plugin{},
		},
		Cmd: exec.Command(file),
	})
	// TODO defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(pluginKey)
	if err != nil {
		return nil, err
	}

	extension, _ := raw.(forge.Forge)
	return extension, nil
}

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "WOODPECKER_PLUGIN",
	MagicCookieValue: "woodpecker-plugin-magic-cookie-value",
}
