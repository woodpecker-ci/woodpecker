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

package stepbuilder

import (
	"fmt"
	"net/url"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/version"
)

// MetadataFromStruct return the metadata from a pipeline will run with.
func MetadataFromStruct(forge metadata.ServerForge, repo *model.Repo, pipeline, prev *model.Pipeline, workflow *model.Workflow, sysURL string) metadata.Metadata {
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
			RemoteID:    fmt.Sprint(repo.ForgeRemoteID),
			ForgeURL:    repo.ForgeURL,
			SCM:         string(repo.SCMKind),
			CloneURL:    repo.Clone,
			CloneSSHURL: repo.CloneSSH,
			Private:     repo.IsSCMPrivate,
			Branch:      repo.Branch,
			Trusted:     repo.IsTrusted,
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
		Curr:     metadataPipelineFromModelPipeline(pipeline, true),
		Prev:     metadataPipelineFromModelPipeline(prev, false),
		Workflow: fWorkflow,
		Step:     metadata.Step{},
		Sys: metadata.System{
			Name:     "woodpecker",
			URL:      sysURL,
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
