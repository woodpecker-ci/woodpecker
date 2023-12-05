package rpc

import (
	"net/rpc"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/shared/addon/types"
)

type AddonRPCClient[T any] struct {
	client *rpc.Client
}

func (a *AddonRPCClient[T]) Type() types.Type {
	var resp types.Type
	err := a.client.Call("Plugin.Type", new(any), &resp)
	if err != nil {
		log.Error().Err(err).Msg("could not get addon type")
		return ""
	}

	return resp
}

func (a *AddonRPCClient[T]) Addon(env []string) (T, error) {
	var resp T
	err := a.client.Call("Plugin.Addon", map[string]any{
		//"logger": logger,
		"env": env,
	}, &resp)

	return resp, err
}
