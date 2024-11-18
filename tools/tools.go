//go:build tools
// +build tools

package main

import (
	_ "github.com/getkin/kin-openapi/cmd/validate"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/swaggo/swag/cmd/swag"
)
