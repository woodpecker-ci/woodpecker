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

type (
	// Metadata defines runtime m.
	Metadata struct {
		ID       string   `json:"id,omitempty"`
		Repo     Repo     `json:"repo,omitempty"`
		Curr     Pipeline `json:"curr,omitempty"`
		Prev     Pipeline `json:"prev,omitempty"`
		Workflow Workflow `json:"workflow,omitempty"`
		Step     Step     `json:"step,omitempty"`
		Sys      System   `json:"sys,omitempty"`
		Forge    Forge    `json:"forge,omitempty"`
	}

	// Repo defines runtime metadata for a repository.
	Repo struct {
		Name     string   `json:"name,omitempty"`
		Owner    string   `json:"owner,omitempty"`
		RemoteID string   `json:"remote_id,omitempty"`
		Link     string   `json:"link,omitempty"`
		CloneURL string   `json:"clone_url,omitempty"`
		Private  bool     `json:"private,omitempty"`
		Secrets  []Secret `json:"secrets,omitempty"`
		Branch   string   `json:"default_branch,omitempty"`
		Trusted  bool     `json:"trusted,omitempty"`
	}

	// Pipeline defines runtime metadata for a pipeline.
	Pipeline struct {
		Number   int64  `json:"number,omitempty"`
		Created  int64  `json:"created,omitempty"`
		Started  int64  `json:"started,omitempty"`
		Finished int64  `json:"finished,omitempty"`
		Timeout  int64  `json:"timeout,omitempty"`
		Status   string `json:"status,omitempty"`
		Event    string `json:"event,omitempty"`
		Link     string `json:"link,omitempty"`
		Target   string `json:"target,omitempty"`
		Trusted  bool   `json:"trusted,omitempty"`
		Commit   Commit `json:"commit,omitempty"`
		Parent   int64  `json:"parent,omitempty"`
		Cron     string `json:"cron,omitempty"`
	}

	// Commit defines runtime metadata for a commit.
	Commit struct {
		Sha               string   `json:"sha,omitempty"`
		Ref               string   `json:"ref,omitempty"`
		Refspec           string   `json:"refspec,omitempty"`
		Branch            string   `json:"branch,omitempty"`
		Message           string   `json:"message,omitempty"`
		Author            Author   `json:"author,omitempty"`
		ChangedFiles      []string `json:"changed_files,omitempty"`
		PullRequestLabels []string `json:"labels,omitempty"`
	}

	// Author defines runtime metadata for a commit author.
	Author struct {
		Name   string `json:"name,omitempty"`
		Email  string `json:"email,omitempty"`
		Avatar string `json:"avatar,omitempty"`
	}

	// Workflow defines runtime metadata for a workflow.
	Workflow struct {
		Name   string            `json:"name,omitempty"`
		Number int               `json:"number,omitempty"`
		Matrix map[string]string `json:"matrix,omitempty"`
	}

	// Step defines runtime metadata for a step.
	Step struct {
		Name   string `json:"name,omitempty"`
		Number int    `json:"number,omitempty"`
	}

	// Secret defines a runtime secret
	Secret struct {
		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
		Mount string `json:"mount,omitempty"`
		Mask  bool   `json:"mask,omitempty"`
	}

	// System defines runtime metadata for a ci/cd system.
	System struct {
		Name     string `json:"name,omitempty"`
		Host     string `json:"host,omitempty"`
		Link     string `json:"link,omitempty"`
		Platform string `json:"arch,omitempty"`
		Version  string `json:"version,omitempty"`
	}

	// Forge defines runtime metadata about the forge that host the repo
	Forge struct {
		Type string `json:"type,omitempty"`
		URL  string `json:"url,omitempty"`
	}

	// ServerForge represent the needed func of a server forge to get its metadata
	ServerForge interface {
		// Name returns the string name of this driver
		Name() string
		// URL returns the root url of a configured forge
		URL() string
	}
)
