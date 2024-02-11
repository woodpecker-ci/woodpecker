package forgeaddon

import (
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/hashicorp"
)

var Addon hashicorp.Plugin[forge.Forge] = &Plugin{}
