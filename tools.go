//go:build tools
// +build tools

// this file make sure tools are vendored too
package tools

import (
	_ "github.com/bmatcuk/doublestar"
	_ "github.com/woodpecker-ci/togo"
)
