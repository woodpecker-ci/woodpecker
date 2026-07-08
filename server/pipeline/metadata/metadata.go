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

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/version"
)

type ServerMetadata struct {
	// base holds every metadata field that is identical for all workflows of
	// a pipeline. It is computed once at construction time; GetWorkflowMetadata
	// only stamps the per-workflow fields onto a copy of it.
	base metadata.Metadata
}

func NewServerMetadata(forge metadata.ServerForge, repo *model.Repo, pipeline, previousPipeline *model.Pipeline, sysURL string) *ServerMetadata {
	host := sysURL
	uri, err := url.Parse(sysURL)
	if err == nil {
		host = uri.Host
	}

	fForge := metadata.Forge{}
	if forge != nil {
		fForge = metadata.Forge{
			Type: forge.Name(),
			URL:  forge.URL(),
		}
	}

	fRepo := metadata.Repo{}
	if repo != nil {
		fRepo = metadata.Repo{
			ID:          repo.ID,
			Name:        repo.Name,
			Owner:       repo.Owner,
			OrgID:       repo.OrgID,
			RemoteID:    fmt.Sprint(repo.ForgeRemoteID),
			ForgeURL:    repo.ForgeURL,
			CloneURL:    repo.Clone,
			CloneSSHURL: repo.CloneSSH,
			Private:     repo.IsSCMPrivate,
			Branch:      repo.Branch,
			Trusted: metadata.TrustedConfiguration{
				Network:  repo.Trusted.Network,
				Volumes:  repo.Trusted.Volumes,
				Security: repo.Trusted.Security,
			},
		}

		if idx := strings.LastIndex(repo.FullName, "/"); idx != -1 {
			if fRepo.Name == "" && repo.FullName != "" {
				fRepo.Name = repo.FullName[idx+1:]
			}
			if fRepo.Owner == "" && repo.FullName != "" {
				fRepo.Owner = repo.FullName[:idx]
			}
		}
	}

	return &ServerMetadata{
		base: metadata.Metadata{
			Repo: fRepo,
			Curr: metadataPipelineFromModelPipeline(pipeline, true),
			Prev: metadataPipelineFromModelPipeline(previousPipeline, false),
			Step: metadata.Step{},
			Sys: metadata.System{
				Name:     "woodpecker",
				URL:      sysURL,
				Host:     host,
				Platform: "", // will be set by pipeline platform option or by agent
				Version:  version.Version,
			},
			Forge: fForge,
		},
	}
}

// GetWorkflowMetadata return the metadata from a pipeline will run with.
// TODO: builder should depend on metadata not the other way around
func (s *ServerMetadata) GetWorkflowMetadata(workflow *builder.Workflow) metadata.Metadata {
	m := s.base

	if workflow != nil {
		m.Workflow = metadata.Workflow{
			Name:   workflow.Name,
			Number: workflow.PID,
			Matrix: workflow.Environ,
		}
	}

	return m
}

func metadataPipelineFromModelPipeline(pipeline *model.Pipeline, includeParent bool) metadata.Pipeline {
	if pipeline == nil {
		return metadata.Pipeline{}
	}

	parent := int64(0)
	if includeParent {
		parent = pipeline.Parent
	}

	metadata := metadata.Pipeline{
		Number:      pipeline.Number,
		Parent:      parent,
		Created:     pipeline.Created,
		Started:     pipeline.Started,
		Finished:    pipeline.Finished,
		Status:      string(pipeline.Status),
		Event:       metadata.Event(pipeline.Event),
		EventReason: pipeline.EventReason,
		ForgeURL:    pipeline.ForgeURL,
		RerunCount:  pipeline.RerunCount,
		DeployTo:    pipeline.DeployTo,
		DeployTask:  pipeline.DeployTask,
		Commit: metadata.Commit{
			Sha:       pipeline.Commit,
			Ref:       pipeline.Ref,
			Refspec:   pipeline.Refspec,
			Branch:    pipeline.Branch,
			Message:   pipeline.Message,
			Timestamp: pipeline.Timestamp,
			Author: metadata.Author{
				Name:  pipeline.Author,
				Email: pipeline.Email,
			},
			ChangedFiles:         pipeline.ChangedFiles,
			PullRequestLabels:    pipeline.PullRequestLabels,
			PullRequestMilestone: pipeline.PullRequestMilestone,
			PullRequestDraft:     pipeline.PullRequestDraft,
		},
		Cron:   pipeline.Cron,
		Author: pipeline.Author,
		Avatar: pipeline.Avatar,
	}

	if pipeline.Release != nil {
		metadata.Release.Title = pipeline.Release.Title
		metadata.Release.IsPrerelease = pipeline.Release.IsPrerelease
	}

	return metadata
}
