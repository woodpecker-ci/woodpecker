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

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/version"
)

type MetadataServerForge struct {
	forge        metadata.ServerForge
	repo         *model.Repo
	pipeline     *model.Pipeline
	prevPipeline *model.Pipeline
	sysURL       string
}

func NewMetadataServerForge(forge metadata.ServerForge, repo *model.Repo, pipeline *model.Pipeline, prevPipeline *model.Pipeline, sysURL string) *MetadataServerForge {
	return &MetadataServerForge{
		forge:        forge,
		repo:         repo,
		pipeline:     pipeline,
		prevPipeline: prevPipeline,
		sysURL:       sysURL,
	}
}

// MetadataForWorkflow returns the metadata for a workflow.
func (m *MetadataServerForge) MetadataForWorkflow(workflow *model.Workflow) metadata.Metadata {
	host := m.sysURL
	uri, err := url.Parse(m.sysURL)
	if err == nil {
		host = uri.Host
	}

	fForge := metadata.Forge{}
	if m.forge != nil {
		fForge = metadata.Forge{
			Type: m.forge.Name(),
			URL:  m.forge.URL(),
		}
	}

	fRepo := metadata.Repo{}
	if m.repo != nil {
		fRepo = metadata.Repo{
			ID:          m.repo.ID,
			Name:        m.repo.Name,
			Owner:       m.repo.Owner,
			RemoteID:    fmt.Sprint(m.repo.ForgeRemoteID),
			ForgeURL:    m.repo.ForgeURL,
			SCM:         string(m.repo.SCMKind),
			CloneURL:    m.repo.Clone,
			CloneSSHURL: m.repo.CloneSSH,
			Private:     m.repo.IsSCMPrivate,
			Branch:      m.repo.Branch,
			Trusted: metadata.TrustedConfiguration{
				Network:  m.repo.Trusted.Network,
				Volumes:  m.repo.Trusted.Volumes,
				Security: m.repo.Trusted.Security,
			},
		}

		if idx := strings.LastIndex(m.repo.FullName, "/"); idx != -1 {
			if fRepo.Name == "" && m.repo.FullName != "" {
				fRepo.Name = m.repo.FullName[idx+1:]
			}
			if fRepo.Owner == "" && m.repo.FullName != "" {
				fRepo.Owner = m.repo.FullName[:idx]
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
		Curr:     metadataPipelineFromModelPipeline(m.pipeline, true),
		Prev:     metadataPipelineFromModelPipeline(m.prevPipeline, false),
		Workflow: fWorkflow,
		Step:     metadata.Step{},
		Sys: metadata.System{
			Name:     "woodpecker",
			URL:      m.sysURL,
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
		Number:     pipeline.Number,
		Parent:     parent,
		Created:    pipeline.Created,
		Started:    pipeline.Started,
		Finished:   pipeline.Finished,
		Status:     string(pipeline.Status),
		Event:      string(pipeline.Event),
		ForgeURL:   pipeline.ForgeURL,
		DeployTo:   pipeline.DeployTo,
		DeployTask: pipeline.DeployTask,
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
			ChangedFiles:      pipeline.ChangedFiles,
			PullRequestLabels: pipeline.PullRequestLabels,
			IsPrerelease:      pipeline.IsPrerelease,
		},
		Cron: cron,
	}
}
