package permissions

import (
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

func NewOrgs(orgs []string) *Orgs {
	orgsLowercase := make([]string, len(orgs))
	for i, a := range orgs {
		orgsLowercase[i] = strings.ToLower(a)
	}
	return &Orgs{
		IsConfigured: len(orgs) > 0,
		orgs:         utils.SliceToBoolMap(orgsLowercase),
	}
}

type Orgs struct {
	IsConfigured bool
	orgs         map[string]bool
}

func (o *Orgs) IsMember(teams []*model.Team) bool {
	for _, team := range teams {
		if o.orgs[strings.ToLower(team.Login)] {
			return true
		}
	}
	return false
}
