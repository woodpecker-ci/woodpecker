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
		ID       string
		Repo     Repo
		Curr     Pipeline
		Prev     Pipeline
		Workflow Workflow
		Step     Step
		Sys      System
		Forge    Forge
	}

	// Repo defines runtime metadata for a repository.
	Repo struct {
		ID          int64
		Name        string
		Owner       string
		RemoteID    string
		ForgeURL    string
		CloneURL    string
		CloneSSHURL string
		Private     bool
		Secrets     []Secret
		Branch      string
		Trusted     bool
	}

	// Pipeline defines runtime metadata for a pipeline.
	Pipeline struct {
		Number   int64
		Created  int64
		Started  int64
		Finished int64
		Timeout  int64
		Status   string
		Event    string
		ForgeURL string
		Target   string
		Trusted  bool
		Commit   Commit
		Parent   int64
		Cron     string
	}

	// Commit defines runtime metadata for a commit.
	Commit struct {
		Sha               string
		Ref               string
		Refspec           string
		Branch            string
		Message           string
		Author            Author
		ChangedFiles      []string
		PullRequestLabels []string
	}

	// Author defines runtime metadata for a commit author.
	Author struct {
		Name   string
		Email  string
		Avatar string
	}

	// Workflow defines runtime metadata for a workflow.
	Workflow struct {
		Name   string
		Number int
		Matrix map[string]string
	}

	// Step defines runtime metadata for a step.
	Step struct {
		Name   string
		Number int
	}

	// Secret defines a runtime secret
	Secret struct {
		Name  string
		Value string
		Mount string
		Mask  bool
	}

	// System defines runtime metadata for a ci/cd system.
	System struct {
		Name     string
		Host     string
		URL      string
		Platform string
		Version  string
	}

	// Forge defines runtime metadata about the forge that host the repo
	Forge struct {
		Type string
		URL  string
	}

	// ServerForge represent the needed func of a server forge to get its metadata
	ServerForge interface {
		// Name returns the string name of this driver
		Name() string
		// URL returns the root url of a configured forge
		URL() string
	}
)
