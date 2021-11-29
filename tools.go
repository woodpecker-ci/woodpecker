//go:build tools
// +build tools

// this file make sure tools are vendored too
package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/rs/zerolog/cmd/lint"
)
