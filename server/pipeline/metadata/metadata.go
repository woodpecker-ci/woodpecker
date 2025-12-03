// Copyright 2023 Woodpecker Authors
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

package metadata

import (
	"fmt"
	"net/url"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/builder"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/version"
)

type ServerMetadata struct {
	forge            metadata.ServerForge
	repo             *model.Repo
	pipeline         *model.Pipeline
	previousPipeline *model.Pipeline
	sysURL           string
}

func NewServerMetadata(forge metadata.ServerForge, repo *model.Repo, pipeline, previousPipeline *model.Pipeline, sysURL string) *ServerMetadata {
	return &ServerMetadata{
		forge:            forge,
		repo:             repo,
		pipeline:         pipeline,
		previousPipeline: previousPipeline,
		sysURL:           sysURL,
	}
}

// GetWorkflowMetadata return the metadata from a pipeline will run with.
func (s *ServerMetadata) GetWorkflowMetadata(workflow *builder.Workflow) metadata.Metadata {
	host := s.sysURL
	uri, err := url.Parse(s.sysURL)
	if err == nil {
		host = uri.Host
	}

	fForge := metadata.Forge{}
	if s.forge != nil {
		fForge = metadata.Forge{
			Type: s.forge.Name(),
			URL:  s.forge.URL(),
		}
	}

	fRepo := metadata.Repo{}
	if s.repo != nil {
		fRepo = metadata.Repo{
			ID:          s.repo.ID,
			Name:        s.repo.Name,
			Owner:       s.repo.Owner,
			RemoteID:    fmt.Sprint(s.repo.ForgeRemoteID),
			ForgeURL:    s.repo.ForgeURL,
			CloneURL:    s.repo.Clone,
			CloneSSHURL: s.repo.CloneSSH,
			Private:     s.repo.IsSCMPrivate,
			Branch:      s.repo.Branch,
			Trusted: metadata.TrustedConfiguration{
				Network:  s.repo.Trusted.Network,
				Volumes:  s.repo.Trusted.Volumes,
				Security: s.repo.Trusted.Security,
			},
		}

		if idx := strings.LastIndex(s.repo.FullName, "/"); idx != -1 {
			if fRepo.Name == "" && s.repo.FullName != "" {
				fRepo.Name = s.repo.FullName[idx+1:]
			}
			if fRepo.Owner == "" && s.repo.FullName != "" {
				fRepo.Owner = s.repo.FullName[:idx]
			}
		}
	}

	fWorkflow := metadata.Workflow{}
	if workflow != nil {
		fWorkflow = metadata.Workflow{
			Name:   workflow.Name,
			Number: workflow.PID,
			Matrix: workflow.Environ,
		}
	}

	return metadata.Metadata{
		Repo:     fRepo,
		Curr:     metadataPipelineFromModelPipeline(s.pipeline, true),
		Prev:     metadataPipelineFromModelPipeline(s.previousPipeline, false),
		Workflow: fWorkflow,
		Step:     metadata.Step{},
		Sys: metadata.System{
			Name:     "woodpecker",
			URL:      s.sysURL,
			Host:     host,
			Platform: "", // will be set by pipeline platform option or by agent
			Version:  version.Version,
		},
		Forge: fForge,
	}
}

func metadataPipelineFromModelPipeline(pipeline *model.Pipeline, includeParent bool) metadata.Pipeline {
	if pipeline == nil {
		return metadata.Pipeline{}
	}

	cron := ""
	if pipeline.Event == model.EventCron {
		cron = pipeline.Sender
	}

	parent := int64(0)
	if includeParent {
		parent = pipeline.Parent
	}

	return metadata.Pipeline{
		Number:      pipeline.Number,
		Parent:      parent,
		Created:     pipeline.Created,
		Started:     pipeline.Started,
		Finished:    pipeline.Finished,
		Status:      string(pipeline.Status),
		Event:       string(pipeline.Event),
		EventReason: pipeline.EventReason,
		ForgeURL:    pipeline.ForgeURL,
		DeployTo:    pipeline.DeployTo,
		DeployTask:  pipeline.DeployTask,
		Commit: metadata.Commit{
			Sha:     pipeline.Commit,
			Ref:     pipeline.Ref,
			Refspec: pipeline.Refspec,
			Branch:  pipeline.Branch,
			Message: pipeline.Message,
			Author: metadata.Author{
				Name:   pipeline.Author,
				Email:  pipeline.Email,
				Avatar: pipeline.Avatar,
			},
			ChangedFiles:         pipeline.ChangedFiles,
			PullRequestLabels:    pipeline.PullRequestLabels,
			PullRequestMilestone: pipeline.PullRequestMilestone,
			IsPrerelease:         pipeline.IsPrerelease,
		},
		Cron:   cron,
		Author: pipeline.Author,
		Avatar: pipeline.Avatar,
	}
}
