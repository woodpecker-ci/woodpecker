package rpc

import (
	//"net/rpc"

	"github.com/hashicorp/go-plugin"

	"go.woodpecker-ci.org/woodpecker/shared/addon/types"
)

type AddonRPCServer[T any] struct {
	Impl   types.Addon[T]
	broker *plugin.MuxBroker
}

func (a *AddonRPCServer[T]) Type(_ interface{}, resp *types.Type) error {
	*resp = a.Impl.Type()
	return nil
}

func (a *AddonRPCServer[T]) Addon(args map[string]interface{}, resp *T) error {
	addon, err := a.Impl.Addon(args["env"].([]string))
	*resp = addon
	return err
}
