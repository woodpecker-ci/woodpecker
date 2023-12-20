package addon

import (
	"errors"
	"os"
	"plugin"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/shared/addon/types"
)

var pluginCache = map[string]*plugin.Plugin{}

type Addon[T any] struct {
	Type  types.Type
	Value T
}

func Load[T any](files []string, t types.Type) (*Addon[T], error) {
	for _, file := range files {
		if _, has := pluginCache[file]; !has {
			p, err := plugin.Open(file)
			if err != nil {
				return nil, err
			}
			pluginCache[file] = p
		}

		typeLookup, err := pluginCache[file].Lookup("Type")
		if err != nil {
			return nil, err
		}
		if addonType, is := typeLookup.(*types.Type); !is {
			return nil, errors.New("addon type is incorrect")
		} else if *addonType != t {
			continue
		}

		mainLookup, err := pluginCache[file].Lookup("Addon")
		if err != nil {
			return nil, err
		}
		main, is := mainLookup.(func(zerolog.Logger, []string) (T, error))
		if !is {
			return nil, errors.New("addon main function has incorrect type")
		}

		logger := log.Logger.With().Str("addon", file).Logger()

		mainOut, err := main(logger, os.Environ())
		if err != nil {
			return nil, err
		}
		return &Addon[T]{
			Type:  t,
			Value: mainOut,
		}, nil
	}

	return nil, nil
}
