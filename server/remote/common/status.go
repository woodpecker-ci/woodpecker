package common

import (
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

const base = "ci/woodpecker"

func GetBuildStatusContext(repo *model.Repo, build *model.Build, proc *model.Proc) string {
	name := base // TODO: use "status-context"

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

// getBuildStatusDescription is a helper function that generates a description
// message for the current build status.
func GetBuildStatusDescription(status model.StatusValue) string {
	switch status {
	case model.StatusPending:
		return "Pipeline is pending"
	case model.StatusRunning:
		return "Pipeline is running"
	case model.StatusSuccess:
		return "Pipeline was successful"
	case model.StatusFailure, model.StatusError:
		return "Pipeline failed"
	case model.StatusKilled:
		return "Pipeline was canceled"
	case model.StatusBlocked:
		return "Pipeline is pending approval"
	case model.StatusDeclined:
		return "Pipeline was rejected"
	default:
		return "unknown status"
	}
}

func GetBuildStatusLink(repo *model.Repo, build *model.Build, proc *model.Proc) string {
	if proc == nil {
		return fmt.Sprintf("%s/%s/build/%d", server.Config.Server.Host, repo.FullName, build.Number)
	}

	return fmt.Sprintf("%s/%s/build/%d/%d", server.Config.Server.Host, repo.FullName, build.Number, proc.PID)
}
