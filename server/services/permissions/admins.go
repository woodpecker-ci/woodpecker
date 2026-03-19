package permissions

import (
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

func NewAdmins(admins []string) *Admins {
	adminsLowercase := make([]string, len(admins))
	for _, a := range admins {
		adminsLowercase = append(adminsLowercase, strings.ToLower(a))
	}
	return &Admins{admins: utils.SliceToBoolMap(admins)}
}

type Admins struct {
	admins map[string]bool
}

func (a *Admins) IsAdmin(user *model.User) bool {
	return a.admins[strings.ToLower(user.Login)]
}
