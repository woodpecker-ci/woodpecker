package permissions

import (
	"go.woodpecker-ci.org/woodpecker/server/model"
	"go.woodpecker-ci.org/woodpecker/shared/utils"
)

func NewOrgs(orgs []string) *Orgs {
	return &Orgs{
		IsConfigured: len(orgs) > 0,
		orgs:         utils.SliceToBoolMap(orgs),
	}
}

type Orgs struct {
	IsConfigured bool
	orgs         map[string]bool
}

func (o *Orgs) IsMember(teams []*model.Team) bool {
	for _, team := range teams {
		if o.orgs[team.Login] {
			return true
		}
	}
	return false
}
