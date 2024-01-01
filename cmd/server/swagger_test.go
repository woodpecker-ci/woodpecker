package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/cmd/server/docs"
)

func TestSetupSwaggerStaticConfig(t *testing.T) {
	setupSwaggerStaticConfig()
	assert.Equal(t, "/api", docs.SwaggerInfo.BasePath)
}
