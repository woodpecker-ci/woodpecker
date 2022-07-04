package common

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func GetBuildStatusContext(repo *model.Repo, build *model.Build, proc *model.Proc) string {
	event := string(build.Event)
	if build.Event == model.EventPull {
		event = "pr"
	}

	tmpl, err := template.New("context").Parse(server.Config.Server.StatusContextFormat)
	if err != nil {
		return ""
	}
	var ctx bytes.Buffer
	err = tmpl.Execute(&ctx, map[string]interface{}{
		"context":  server.Config.Server.StatusContext,
		"event":    event,
		"pipeline": proc.Name,
		"owner":    repo.Owner,
		"repo":     repo.Name,
	})
	if err != nil {
		return ""
	}

	return ctx.String()
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
