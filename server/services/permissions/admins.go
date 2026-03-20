package permissions

import (
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

func NewAdmins(admins []string) *Admins {
	adminsLowercase := make([]string, len(admins))
	for i, a := range admins {
		adminsLowercase[i] = strings.ToLower(a)
	}
	return &Admins{admins: utils.SliceToBoolMap(adminsLowercase)}
}

type Admins struct {
	admins map[string]bool
}

func (a *Admins) IsAdmin(user *model.User) bool {
	return a.admins[strings.ToLower(user.Login)]
}
