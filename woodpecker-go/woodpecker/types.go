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

package woodpecker

type (
	// User represents a user account.
	User struct {
		ID     int64  `json:"id"`
		Login  string `json:"login"`
		Email  string `json:"email"`
		Avatar string `json:"avatar_url"`
		Active bool   `json:"active"`
		Admin  bool   `json:"admin"`
	}

	// Repo represents a repository.
	Repo struct {
		ID         int64  `json:"id,omitempty"`
		Owner      string `json:"owner"`
		Name       string `json:"name"`
		FullName   string `json:"full_name"`
		Avatar     string `json:"avatar_url,omitempty"`
		Link       string `json:"link_url,omitempty"`
		Kind       string `json:"scm,omitempty"`
		Clone      string `json:"clone_url,omitempty"`
		Branch     string `json:"default_branch,omitempty"`
		Timeout    int64  `json:"timeout,omitempty"`
		Visibility string `json:"visibility"`
		IsPrivate  bool   `json:"private,omitempty"`
		IsTrusted  bool   `json:"trusted"`
		IsStarred  bool   `json:"starred,omitempty"`
		IsGated    bool   `json:"gated"`
		AllowPull  bool   `json:"allow_pr"`
		Config     string `json:"config_file"`
	}

	// RepoPatch defines a repository patch request.
	RepoPatch struct {
		Config          *string `json:"config_file,omitempty"`
		IsTrusted       *bool   `json:"trusted,omitempty"`
		IsGated         *bool   `json:"gated,omitempty"`
		Timeout         *int64  `json:"timeout,omitempty"`
		Visibility      *string `json:"visibility"`
		AllowPull       *bool   `json:"allow_pr,omitempty"`
		PipelineCounter *int    `json:"pipeline_counter,omitempty"`
	}

	// Pipeline defines a pipeline object.
	Pipeline struct {
		ID        int64   `json:"id"`
		Number    int     `json:"number"`
		Parent    int     `json:"parent"`
		Event     string  `json:"event"`
		Status    string  `json:"status"`
		Error     string  `json:"error"`
		Enqueued  int64   `json:"enqueued_at"`
		Created   int64   `json:"created_at"`
		Started   int64   `json:"started_at"`
		Finished  int64   `json:"finished_at"`
		Deploy    string  `json:"deploy_to"`
		Commit    string  `json:"commit"`
		Branch    string  `json:"branch"`
		Ref       string  `json:"ref"`
		Refspec   string  `json:"refspec"`
		CloneURL  string  `json:"clone_url"`
		Title     string  `json:"title"`
		Message   string  `json:"message"`
		Timestamp int64   `json:"timestamp"`
		Sender    string  `json:"sender"`
		Author    string  `json:"author"`
		Avatar    string  `json:"author_avatar"`
		Email     string  `json:"author_email"`
		Link      string  `json:"link_url"`
		Reviewer  string  `json:"reviewed_by"`
		Reviewed  int64   `json:"reviewed_at"`
		Procs     []*Proc `json:"procs,omitempty"`
	}

	// Proc represents a process in the pipeline.
	Proc struct {
		ID       int64             `json:"id"`
		PID      int               `json:"pid"`
		PPID     int               `json:"ppid"`
		PGID     int               `json:"pgid"`
		Name     string            `json:"name"`
		State    string            `json:"state"`
		Error    string            `json:"error,omitempty"`
		ExitCode int               `json:"exit_code"`
		Started  int64             `json:"start_time,omitempty"`
		Stopped  int64             `json:"end_time,omitempty"`
		Machine  string            `json:"machine,omitempty"`
		Platform string            `json:"platform,omitempty"`
		Environ  map[string]string `json:"environ,omitempty"`
		Children []*Proc           `json:"children,omitempty"`
	}

	// Registry represents a docker registry with credentials.
	Registry struct {
		ID       int64  `json:"id"`
		Address  string `json:"address"`
		Username string `json:"username"`
		Password string `json:"password,omitempty"`
		Email    string `json:"email"`
		Token    string `json:"token"`
	}

	// Secret represents a secret variable, such as a password or token.
	Secret struct {
		ID          int64    `json:"id"`
		Name        string   `json:"name"`
		Value       string   `json:"value,omitempty"`
		Images      []string `json:"image"`
		PluginsOnly bool     `json:"plugins_only"`
		Events      []string `json:"event"`
	}

	// Activity represents an item in the user's feed or timeline.
	Activity struct {
		Owner    string `json:"owner"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Number   int    `json:"number,omitempty"`
		Event    string `json:"event,omitempty"`
		Status   string `json:"status,omitempty"`
		Created  int64  `json:"created_at,omitempty"`
		Started  int64  `json:"started_at,omitempty"`
		Finished int64  `json:"finished_at,omitempty"`
		Commit   string `json:"commit,omitempty"`
		Branch   string `json:"branch,omitempty"`
		Ref      string `json:"ref,omitempty"`
		Refspec  string `json:"refspec,omitempty"`
		CloneURL string `json:"clone_url,omitempty"`
		Title    string `json:"title,omitempty"`
		Message  string `json:"message,omitempty"`
		Author   string `json:"author,omitempty"`
		Avatar   string `json:"author_avatar,omitempty"`
		Email    string `json:"author_email,omitempty"`
	}

	// Version provides system version details.
	Version struct {
		Source  string `json:"source,omitempty"`
		Version string `json:"version,omitempty"`
		Commit  string `json:"commit,omitempty"`
	}

	// Info provides queue stats.
	Info struct {
		Stats struct {
			Workers       int `json:"worker_count"`
			Pending       int `json:"pending_count"`
			WaitingOnDeps int `json:"waiting_on_deps_count"`
			Running       int `json:"running_count"`
			Complete      int `json:"completed_count"`
		} `json:"stats"`
		Paused bool `json:"paused,omitempty"`
	}

	// LogLevel is for checking/setting logging level
	LogLevel struct {
		Level string `json:"log-level"`
	}

	// Logs is the JSON data for a logs response
	Logs struct {
		Proc   string `json:"proc"`
		Output string `json:"out"`
	}

	// Cron is the JSON data of a cron job
	Cron struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		RepoID    int64  `json:"repo_id"`
		CreatorID int64  `json:"creator_id"`
		NextExec  int64  `json:"next_exec"`
		Schedule  string `json:"schedule"`
		Created   int64  `json:"created_at"`
		Branch    string `json:"branch"`
	}

	// PipelineOptions is the JSON data for creating a new pipeline
	PipelineOptions struct {
		Branch    string            `json:"branch"`
		Variables map[string]string `json:"variables"`
	}
)
