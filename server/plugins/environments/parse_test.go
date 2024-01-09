package environments

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	service := Parse([]string{})
	env, err := service.EnvironList(nil)
	assert.NoError(t, err)
	assert.Empty(t, env)

	service = Parse([]string{"ENV:value"})
	env, err = service.EnvironList(nil)
	assert.NoError(t, err)
	assert.Len(t, env, 1)
	assert.Equal(t, env[0].Name, "ENV")
	assert.Equal(t, env[0].Value, "value")

	service = Parse([]string{"ENV:value", "ENV2:value2"})
	env, err = service.EnvironList(nil)
	assert.NoError(t, err)
	assert.Len(t, env, 2)

	service = Parse([]string{"ENV:value", "ENV2:value2", "ENV3_WITHOUT_VALUE"})
	env, err = service.EnvironList(nil)
	assert.NoError(t, err)
	assert.Len(t, env, 2)
}
