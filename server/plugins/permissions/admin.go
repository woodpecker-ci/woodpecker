package permissions

import (
	"go.woodpecker-ci.org/woodpecker/server/model"
	"go.woodpecker-ci.org/woodpecker/shared/utils"
)

func NewAdmins(admins []string) *Admins {
	return &Admins{admins: utils.SliceToBoolMap(admins)}
}

type Admins struct {
	admins map[string]bool
}

func (a *Admins) IsAdmin(user *model.User) bool {
	return a.admins[user.Login]
}
