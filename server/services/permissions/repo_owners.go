package permissions

import (
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

func NewOwnersAllowlist(owners []string) *OwnersAllowlist {
	ownersLowercase := make([]string, len(owners))
	for _, a := range owners {
		ownersLowercase = append(ownersLowercase, strings.ToLower(a))
	}
	return &OwnersAllowlist{owners: utils.SliceToBoolMap(ownersLowercase)}
}

type OwnersAllowlist struct {
	owners map[string]bool
}

func (o *OwnersAllowlist) IsAllowed(repo *model.Repo) bool {
	return len(o.owners) < 1 || o.owners[strings.ToLower(repo.Owner)]
}
