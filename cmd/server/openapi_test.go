package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/cmd/server/openapi"
)

func TestSetupOpenApiStaticConfig(t *testing.T) {
	setupOpenAPIStaticConfig()
	assert.Equal(t, "/api", openapi.SwaggerInfo.BasePath)
}
