package configservice

import (
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/shared/addon/hashicorp"
)

var Addon hashicorp.Plugin[forge.Forge] = &Plugin{}
