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

package frontend

import (
	"net/url"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/metadata"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/version"
)

// MetadataFromStruct return the metadata from a pipeline will run with.
func MetadataFromStruct(forge metadata.ServerForge, repo *model.Repo, pipeline, last *model.Pipeline, workflow *model.Step, link string) metadata.Metadata {
	host := link
	uri, err := url.Parse(link)
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
			Name:     repo.FullName,
			Link:     repo.Link,
			CloneURL: repo.Clone,
			Private:  repo.IsSCMPrivate,
			Branch:   repo.Branch,
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
		Prev:     metadataPipelineFromModelPipeline(last, false),
		Workflow: fWorkflow,
		Step:     metadata.Step{},
		Sys: metadata.System{
			Name:     "woodpecker",
			Link:     link,
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
		Number:   pipeline.Number,
		Parent:   parent,
		Created:  pipeline.Created,
		Started:  pipeline.Started,
		Finished: pipeline.Finished,
		Status:   string(pipeline.Status),
		Event:    string(pipeline.Event),
		Link:     pipeline.Link,
		Target:   pipeline.Deploy,
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
		},
		Cron: cron,
	}
}
