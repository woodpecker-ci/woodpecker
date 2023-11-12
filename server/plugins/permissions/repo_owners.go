package permissions

import (
	"go.woodpecker-ci.org/woodpecker/server/model"
	"go.woodpecker-ci.org/woodpecker/shared/utils"
)

func NewOwnersAllowlist(owners []string) *OwnersAllowlist {
	return &OwnersAllowlist{owners: utils.SliceToBoolMap(owners)}
}

type OwnersAllowlist struct {
	owners map[string]bool
}

func (o *OwnersAllowlist) IsAllowed(repo *model.Repo) bool {
	return len(o.owners) > 0 && o.owners[repo.Owner]
}
