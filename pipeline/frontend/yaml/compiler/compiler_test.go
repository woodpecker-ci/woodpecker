package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	yaml_types "github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
	yaml_base_types "github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types/base"
)

func TestSecretAvailable(t *testing.T) {
	secret := Secret{
		Match:      []string{"golang"},
		PluginOnly: false,
	}
	assert.True(t, secret.Available(&yaml_types.Container{
		Image:    "golang",
		Commands: yaml_base_types.StringOrSlice{"echo 'this is not a plugin'"},
	}))
	assert.False(t, secret.Available(&yaml_types.Container{
		Image:    "not-golang",
		Commands: yaml_base_types.StringOrSlice{"echo 'this is not a plugin'"},
	}))
	// secret only available for "golang" plugin
	secret = Secret{
		Match:      []string{"golang"},
		PluginOnly: true,
	}
	assert.True(t, secret.Available(&yaml_types.Container{
		Image:    "golang",
		Commands: yaml_base_types.StringOrSlice{},
	}))
	assert.False(t, secret.Available(&yaml_types.Container{
		Image:    "not-golang",
		Commands: yaml_base_types.StringOrSlice{},
	}))
	assert.False(t, secret.Available(&yaml_types.Container{
		Image:    "not-golang",
		Commands: yaml_base_types.StringOrSlice{"echo 'this is not a plugin'"},
	}))
}

func TestCompilerCompile(t *testing.T) {

}
