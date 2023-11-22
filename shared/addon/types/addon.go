package types

type Addon[T any] interface {
	Type() Type
	Addon([]string) (T, error)
}
