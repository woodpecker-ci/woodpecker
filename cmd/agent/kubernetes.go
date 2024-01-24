//go:build kubernetes

package main

import "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/local"

var backend = local.New()
