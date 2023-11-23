package addon

import (
	"errors"
	"os"
	"os/exec"

	"github.com/hashicorp/go-plugin"

	"go.woodpecker-ci.org/woodpecker/shared/addon/rpc"
	"go.woodpecker-ci.org/woodpecker/shared/addon/types"
)

type Addon[T any] struct {
	Type  types.Type
	Value T
}

func Load[T any](files []string, t types.Type) (*Addon[T], error) {
	//for _, file := range files {
	c := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: rpc.HandshakeConfig,
		Plugins: plugin.PluginSet{
			string(t): &rpc.AddonPlugin[T]{},
		},
		Cmd:    exec.Command(files[0]),
		Logger: nil, // TODO zerolog wrapper
	})

	rpcClient, err := c.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(string(t))
	if err != nil {
		return nil, err
	}

	addon, ok := raw.(types.Addon[T])
	if !ok {
		return nil, errors.New("addon has bad type")
	}

	if addon.Type() != t {
		//continue
		return nil, nil
	}

	mainOut, err := addon.Addon(os.Environ())
	if err != nil {
		return nil, err
	}

	//mainOutTyped, is := mainOut.(T)
	//if !is {
	//return nil, errors.New("main output has bad type")
	//}

	return &Addon[T]{
		Type:  t,
		Value: mainOut,
	}, nil
	//}

	//return nil, nil
}
