// Copyright 2024 Woodpecker Authors
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

package feedback

import (
	"errors"
	"fmt"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"

	cicd_feedback "github.com/6543/cicd_feedback"
)

var (
	ErrNonConvertableStatus = errors.New("got non convertable status")
	ErrWorkflowDepResolve   = errors.New("could not resolve workflow dependency")
	ErrStepDepResolve       = errors.New("could not resolve step dependency")
)

func Convert(in *model.Pipeline, workflows []*model.Workflow) (*cicd_feedback.PipelineResponse, error) {
	pipeStatus, err := convertStatus(in.Status)
	if err != nil {
		return nil, err
	}

	feedbackWorkflows := make([]cicd_feedback.Workflow, 0, len(workflows))
	for _, w := range workflows {
		fw, err := convertWorkflow(*w, in.RepoID, in.Number)
		if err != nil {
			return nil, err
		}
		feedbackWorkflows = append(feedbackWorkflows, *fw)
	}

	out := &cicd_feedback.PipelineResponse{
		PipelineID:           fmt.Sprintf("pipeline_%d", in.ID),
		Title:                in.Title,
		Status:               pipeStatus,
		RequiresManualAction: (in.Status == model.StatusBlocked),
		Workflows:            feedbackWorkflows,
		ExternalURI:          fmt.Sprintf("%s/repos/%d/pipeline/%d", server.Config.Server.Host, in.RepoID, in.Number),
	}

	return out, nil
}

func convertWorkflow(workflow model.Workflow, repoID, pipelineNumber int64) (*cicd_feedback.Workflow, error) {
	feedbackSteps, err := convertSteps(workflow.Children, repoID, pipelineNumber)
	if err != nil {
		return nil, err
	}
	status, err := convertStatus(workflow.State)
	if err != nil {
		return nil, err
	}
	return &cicd_feedback.Workflow{
		ID:     fmt.Sprintf("workflow_%d", workflow.ID),
		Name:   workflow.Name,
		Steps:  feedbackSteps,
		Status: status,
	}, nil
}

func convertSteps(in []*model.Step, repoID, pipelineNumber int64) ([]cicd_feedback.Step, error) {
	result := make([]cicd_feedback.Step, 0, len(in))

	nameUUIDMap := make(map[string]string, len(in))
	for _, s := range in {
		nameUUIDMap[s.Name] = s.UUID
	}

	for _, s := range in {
		status, err := convertStatus(s.State)
		if err != nil {
			return nil, err
		}

		deps := []string{}
		/* TODO: Dependencies
		deps := make([]string, 0, len(s.Dependencies))
		for _, dep := range s.Dependencies {
			uuid, exist := nameUUIDMap[dep]
			if !exist {
				return nil, ErrStepDepResolve
			}
			deps = append(deps, uuid)
		}
		*/

		result = append(result,
			cicd_feedback.Step{
				ID:     s.UUID,
				Name:   s.Name,
				Status: status,
				Inputs: cicd_feedback.Inputs{
					Commands:    nil, // TODO,
					Environment: nil, // TODO
				},
				Outputs: cicd_feedback.Outputs{
					Logs: []cicd_feedback.Log{{
						Name: "console",
						URI:  fmt.Sprintf("%s/api/feedback/%d/%d/%d", server.Config.Server.Host, repoID, pipelineNumber, s.ID),
					}},
				},
				Dependencies: deps,
			})
	}

	return result, nil
}

func PipelineURL(in *model.Pipeline) string {
	return fmt.Sprintf("%s/api/feedback/%d/%d", server.Config.Server.Host, in.RepoID, in.Number)
}

func convertLogs(logs []*model.LogEntry) (string, error) {
	builder := strings.Builder{}
	for _, log := range logs {
		builder.Write(log.Data)
		builder.WriteString("\n")
	}

	return builder.String(), nil
}

func convertStatus(in model.StatusValue) (cicd_feedback.Status, error) {
	switch in {
	case model.StatusPending, model.StatusCreated:
		return cicd_feedback.StatusPending, nil
	case model.StatusSkipped:
		return cicd_feedback.StatusSkipped, nil
	case model.StatusRunning:
		return cicd_feedback.StatusRunning, nil
	case model.StatusSuccess:
		return cicd_feedback.StatusSuccess, nil
	case model.StatusFailure, model.StatusError:
		return cicd_feedback.StatusFailed, nil
	case model.StatusKilled:
		return cicd_feedback.StatusKilled, nil
	case model.StatusBlocked:
		return cicd_feedback.StatusManual, nil
	case model.StatusDeclined:
		return cicd_feedback.StatusDeclined, nil
	default:
		return "", ErrNonConvertableStatus
	}
}
