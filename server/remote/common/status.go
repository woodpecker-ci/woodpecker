package common

import (
	"fmt"
	"strings"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func GetBuildStatusContext(repo *model.Repo, build *model.Build, proc *model.Proc) string {
	ctx := server.Config.Server.StatusContextFormat
	// replace context
	ctx = strings.ReplaceAll(ctx, "%context", server.Config.Server.StatusContext)
	// replace event
	event := string(build.Event)
	switch build.Event {
	case model.EventPull:
		event = "pr"
	}
	ctx = strings.ReplaceAll(ctx, "%event", event)
	// replace pipeline name
	ctx = strings.ReplaceAll(ctx, "%pipeline", proc.Name)
	// replace repo
	ctx = strings.ReplaceAll(ctx, "%owner", repo.Owner)
	ctx = strings.ReplaceAll(ctx, "%repo", repo.Name)
	return ctx
}

// GetBuildStatusDescription is a helper function that generates a description
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
