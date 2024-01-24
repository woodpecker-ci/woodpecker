package configservice

import (
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/config"
	"go.woodpecker-ci.org/woodpecker/v2/shared/addon/hashicorp"
)

var Addon hashicorp.Plugin[config.Extension] = &ExtensionPlugin{}
