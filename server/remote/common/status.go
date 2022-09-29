package common

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func GetPipelineStatusContext(repo *model.Repo, pipeline *model.Pipeline, proc *model.Proc) string {
	event := string(pipeline.Event)
	switch pipeline.Event {
	case model.EventPull:
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

// GetPipelineStatusDescription is a helper function that generates a description
// message for the current pipeline status.
func GetPipelineStatusDescription(status model.StatusValue) string {
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

func GetPipelineStatusLink(repo *model.Repo, pipeline *model.Pipeline, proc *model.Proc) string {
	if proc == nil {
		return fmt.Sprintf("%s/%s/pipeline/%d", server.Config.Server.Host, repo.FullName, pipeline.Number)
	}

	return fmt.Sprintf("%s/%s/pipeline/%d/%d", server.Config.Server.Host, repo.FullName, pipeline.Number, proc.PID)
}
