package common

import "github.com/woodpecker-ci/woodpecker/server/model"

const base = "ci/woodpecker"

func GetStatusName(repo *model.Repo, build *model.Build, proc *model.Proc) string {
	name := base

	switch build.Event {
	case model.EventPull:
		name += "/pr"
	default:
		if len(build.Event) > 0 {
			name += "/" + string(build.Event)
		}
	}

	if proc != nil {
		name += "/" + proc.Name
	}

	return name
}
