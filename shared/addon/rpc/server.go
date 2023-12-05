package rpc

import (
	"go.woodpecker-ci.org/woodpecker/shared/addon/types"
)

type AddonRPCServer[T any] struct {
	Impl types.Addon[T]
}

func (a *AddonRPCServer[T]) Type(_ interface{}, resp *types.Type) error {
	*resp = a.Impl.Type()
	return nil
}

func (a *AddonRPCServer[T]) Addon(args map[string]interface{}, resp *T) error {
	addon, err := a.Impl.Addon( /*args["logger"].(zerolog.Logger), */ args["env"].([]string))
	*resp = addon
	return err
}
