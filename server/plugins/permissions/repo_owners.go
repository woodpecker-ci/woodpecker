package permissions

import "go.woodpecker-ci.org/woodpecker/server/model"

func NewOwnersWhitelist(owners []string) *OwnersWhitelist {
	return &OwnersWhitelist{owners: sliceToMap(owners)}
}

type OwnersWhitelist struct {
	owners map[string]bool
}

func (o *OwnersWhitelist) IsAllowed(repo *model.Repo) bool {
	return len(o.owners) > 0 && o.owners[repo.Owner]
}
