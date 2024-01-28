package secretservice

import (
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/shared/addon/hashicorp"
)

var Addon hashicorp.Plugin[model.SecretService] = &Plugin{}
