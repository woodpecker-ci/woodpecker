package compiler

import (
	"testing"

	"github.com/docker/docker/api/types/strslice"
	"github.com/stretchr/testify/assert"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

func TestSecretAvailable(t *testing.T) {
	secret := Secret{
		Match:      []string{"golang"},
		PluginOnly: false,
	}
	assert.True(t, secret.Available(&yaml.Container{
		Image:    "golang",
		Commands: types.StringOrSlice(strslice.StrSlice{"echo 'this is not a plugin'"}),
	}))
	assert.False(t, secret.Available(&yaml.Container{
		Image:    "not-golang",
		Commands: types.StringOrSlice(strslice.StrSlice{"echo 'this is not a plugin'"}),
	}))
	// secret only available for "golang" plugin
	secret = Secret{
		Match:      []string{"golang"},
		PluginOnly: true,
	}
	assert.True(t, secret.Available(&yaml.Container{
		Image:    "golang",
		Commands: types.StringOrSlice(strslice.StrSlice{}),
	}))
	assert.False(t, secret.Available(&yaml.Container{
		Image:    "not-golang",
		Commands: types.StringOrSlice(strslice.StrSlice{}),
	}))
	assert.False(t, secret.Available(&yaml.Container{
		Image:    "not-golang",
		Commands: types.StringOrSlice(strslice.StrSlice{"echo 'this is not a plugin'"}),
	}))
}
