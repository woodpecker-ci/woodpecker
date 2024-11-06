package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/cmd/server/openapi"
)

func TestSetupOpenApiStaticConfig(t *testing.T) {
	setupOpenApiStaticConfig()
	assert.Equal(t, "/api", openapi.SwaggerInfo.BasePath)
}
