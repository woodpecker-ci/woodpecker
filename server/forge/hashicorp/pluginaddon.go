package hashicorp

import "github.com/hashicorp/go-plugin"

type Plugin[T any] interface {
	plugin.Plugin
	Key() string
	WithImpl(T) plugin.Plugin
}
