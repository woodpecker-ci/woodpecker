//go:build tools
// +build tools

package woodpeckergo

import (
	_ "github.com/getkin/kin-openapi/cmd/validate"
	_ "github.com/swaggo/swag/cmd/swag"
)
