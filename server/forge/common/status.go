// Copyright 2022 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func GetPipelineStatusContext(repo *model.Repo, pipeline *model.Pipeline, workflow *model.Workflow) string {
	event := string(pipeline.Event)
	if pipeline.Event == model.EventPull {
		event = "pr"
	}

	tmpl, err := template.New("context").Parse(server.Config.Server.StatusContextFormat)
	if err != nil {
		log.Error().Err(err).Msg("could not create status from template")
		return ""
	}
	var ctx bytes.Buffer
	err = tmpl.Execute(&ctx, map[string]any{
		"context":  server.Config.Server.StatusContext,
		"event":    event,
		"workflow": workflow.Name,
		"owner":    repo.Owner,
		"repo":     repo.Name,
		"axis_id":  workflow.AxisID,
	})
	if err != nil {
		log.Error().Err(err).Msg("could not create status context")
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

func GetPipelineStatusURL(repo *model.Repo, pipeline *model.Pipeline, workflow *model.Workflow) string {
	if workflow == nil {
		return fmt.Sprintf("%s/repos/%d/pipeline/%d", server.Config.Server.Host, repo.ID, pipeline.Number)
	}

	return fmt.Sprintf("%s/repos/%d/pipeline/%d/%d", server.Config.Server.Host, repo.ID, pipeline.Number, workflow.PID)
}
