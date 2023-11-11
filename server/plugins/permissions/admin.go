package permissions

import "go.woodpecker-ci.org/woodpecker/server/model"

func NewAdmins(admins []string) *Admins {
	return &Admins{admins: sliceToMap(admins)}
}

type Admins struct {
	admins map[string]bool
}

func (a *Admins) IsAdmin(user *model.User) bool {
	return a.admins[user.Login]
}
