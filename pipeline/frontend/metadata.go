package frontend

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/woodpecker-ci/woodpecker/version"
)

// Event types corresponding to scm hooks.
const (
	EventPush   = "push"
	EventPull   = "pull_request"
	EventTag    = "tag"
	EventDeploy = "deployment"
)

type (
	// Metadata defines runtime m.
	Metadata struct {
		ID   string `json:"id,omitempty"`
		Repo Repo   `json:"repo,omitempty"`
		Curr Build  `json:"curr,omitempty"`
		Prev Build  `json:"prev,omitempty"`
		Job  Job    `json:"job,omitempty"`
		Sys  System `json:"sys,omitempty"`
	}

	// Repo defines runtime metadata for a repository.
	Repo struct {
		Name    string   `json:"name,omitempty"`
		Link    string   `json:"link,omitempty"`
		Remote  string   `json:"remote,omitempty"`
		Private bool     `json:"private,omitempty"`
		Secrets []Secret `json:"secrets,omitempty"`
		Branch  string   `json:"default_branch,omitempty"`
	}

	// Build defines runtime metadata for a build.
	Build struct {
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
	}

	// Commit defines runtime metadata for a commit.
	Commit struct {
		Sha          string   `json:"sha,omitempty"`
		Ref          string   `json:"ref,omitempty"`
		Refspec      string   `json:"refspec,omitempty"`
		Branch       string   `json:"branch,omitempty"`
		Message      string   `json:"message,omitempty"`
		Author       Author   `json:"author,omitempty"`
		ChangedFiles []string `json:"changed_files,omitempty"`
	}

	// Author defines runtime metadata for a commit author.
	Author struct {
		Name   string `json:"name,omitempty"`
		Email  string `json:"email,omitempty"`
		Avatar string `json:"avatar,omitempty"`
	}

	// Job defines runtime metadata for a job.
	Job struct {
		Number int               `json:"number,omitempty"`
		Matrix map[string]string `json:"matrix,omitempty"`
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
		Name    string `json:"name,omitempty"`
		Host    string `json:"host,omitempty"`
		Link    string `json:"link,omitempty"`
		Arch    string `json:"arch,omitempty"`
		Version string `json:"version,omitempty"`
	}
)

// Environ returns the metadata as a map of environment variables.
func (m *Metadata) Environ() map[string]string {
	var (
		repoOwner    string
		repoName     string
		sourceBranch string
		targetBranch string
	)

	repoParts := strings.Split(m.Repo.Name, "/")
	if len(repoParts) == 2 {
		repoOwner = repoParts[0]
		repoName = repoParts[1]
	} else {
		repoName = m.Repo.Name
	}

	branchParts := strings.Split(m.Curr.Commit.Refspec, ":")
	if len(branchParts) == 2 {
		sourceBranch = branchParts[0]
		targetBranch = branchParts[1]
	}

	params := map[string]string{
		"CI":                     m.Sys.Name,
		"CI_REPO":                m.Repo.Name,
		"CI_REPO_OWNER":          repoOwner,
		"CI_REPO_NAME":           repoName,
		"CI_REPO_SCM":            "git",
		"CI_REPO_LINK":           m.Repo.Link,
		"CI_REPO_REMOTE":         m.Repo.Remote,
		"CI_REPO_DEFAULT_BRANCH": m.Repo.Branch,
		"CI_REPO_PRIVATE":        strconv.FormatBool(m.Repo.Private),
		"CI_REPO_TRUSTED":        "false", // TODO should this be added?

		"CI_COMMIT_SHA":           m.Curr.Commit.Sha,
		"CI_COMMIT_REF":           m.Curr.Commit.Ref,
		"CI_COMMIT_REFSPEC":       m.Curr.Commit.Refspec,
		"CI_COMMIT_BRANCH":        m.Curr.Commit.Branch,
		"CI_COMMIT_SOURCE_BRANCH": sourceBranch,
		"CI_COMMIT_TARGET_BRANCH": targetBranch,
		"CI_COMMIT_LINK":          m.Curr.Link,
		"CI_COMMIT_MESSAGE":       m.Curr.Commit.Message,
		"CI_COMMIT_AUTHOR":        m.Curr.Commit.Author.Name,
		"CI_COMMIT_AUTHOR_EMAIL":  m.Curr.Commit.Author.Email,
		"CI_COMMIT_AUTHOR_AVATAR": m.Curr.Commit.Author.Avatar,
		"CI_COMMIT_TAG":           "", // will be set if event is tag
		"CI_COMMIT_PULL_REQUEST":  "", // will be set if event is pr

		"CI_BUILD_NUMBER":        strconv.FormatInt(m.Curr.Number, 10),
		"CI_BUILD_PARENT":        strconv.FormatInt(m.Curr.Parent, 10),
		"CI_BUILD_EVENT":         m.Curr.Event,
		"CI_BUILD_LINK":          m.Curr.Link,
		"CI_BUILD_DEPLOY_TARGET": m.Curr.Target,
		"CI_BUILD_STATUS":        m.Curr.Status,
		"CI_BUILD_CREATED":       strconv.FormatInt(m.Curr.Created, 10),
		"CI_BUILD_STARTED":       strconv.FormatInt(m.Curr.Started, 10),
		"CI_BUILD_FINISHED":      strconv.FormatInt(m.Curr.Finished, 10),

		"CI_JOB_NUMBER":   strconv.Itoa(m.Job.Number),
		"CI_JOB_STATUS":   "", // will be set by agent
		"CI_JOB_STARTED":  "", // will be set by agent
		"CI_JOB_FINISHED": "", // will be set by agent

		"CI_PREV_COMMIT_SHA":           m.Prev.Commit.Sha,
		"CI_PREV_COMMIT_REF":           m.Prev.Commit.Ref,
		"CI_PREV_COMMIT_REFSPEC":       m.Prev.Commit.Refspec,
		"CI_PREV_COMMIT_BRANCH":        m.Prev.Commit.Branch,
		"CI_PREV_COMMIT_LINK":          m.Prev.Link,
		"CI_PREV_COMMIT_MESSAGE":       m.Prev.Commit.Message,
		"CI_PREV_COMMIT_AUTHOR":        m.Prev.Commit.Author.Name,
		"CI_PREV_COMMIT_AUTHOR_EMAIL":  m.Prev.Commit.Author.Email,
		"CI_PREV_COMMIT_AUTHOR_AVATAR": m.Prev.Commit.Author.Avatar,

		"CI_PREV_BUILD_NUMBER":        strconv.FormatInt(m.Prev.Number, 10),
		"CI_PREV_BUILD_PARENT":        strconv.FormatInt(m.Prev.Parent, 10),
		"CI_PREV_BUILD_EVENT":         m.Prev.Event,
		"CI_PREV_BUILD_LINK":          m.Prev.Link,
		"CI_PREV_BUILD_DEPLOY_TARGET": m.Prev.Target,
		"CI_PREV_BUILD_STATUS":        m.Prev.Status,
		"CI_PREV_BUILD_CREATED":       strconv.FormatInt(m.Prev.Created, 10),
		"CI_PREV_BUILD_STARTED":       strconv.FormatInt(m.Prev.Started, 10),
		"CI_PREV_BUILD_FINISHED":      strconv.FormatInt(m.Prev.Finished, 10),

		"CI_SYSTEM_NAME":    m.Sys.Name,
		"CI_SYSTEM_LINK":    m.Sys.Link,
		"CI_SYSTEM_HOST":    m.Sys.Host,
		"CI_SYSTEM_ARCH":    m.Sys.Arch,
		"CI_SYSTEM_VERSION": version.Version,

		// TODO drop for v0.16.0 development
		// DEPRECATED
		"CI_ARCH":                    m.Sys.Arch,                           // use CI_SYSTEM_ARCH
		"CI_COMMIT":                  m.Curr.Commit.Sha,                    // use CI_COMMIT_SHA
		"CI_REMOTE_URL":              m.Repo.Remote,                        // use CI_REPO_REMOTE
		"CI_REPO_BRANCH":             m.Repo.Branch,                        // use CI_REPO_DEFAULT_BRANCH
		"CI_PARENT_BUILD_NUMBER":     strconv.FormatInt(m.Curr.Parent, 10), // use CI_BUILD_PARENT
		"CI_BUILD_TARGET":            m.Curr.Target,                        // use CI_BUILD_DEPLOY_TARGET
		"CI_DEPLOY_TO":               m.Curr.Target,                        // use CI_BUILD_DEPLOY_TARGET
		"CI_COMMIT_AUTHOR_NAME":      m.Curr.Commit.Author.Name,            // use CI_COMMIT_AUTHOR
		"CI_PREV_COMMIT_AUTHOR_NAME": m.Prev.Commit.Author.Name,            // use CI_PREV_COMMIT_AUTHOR
		"CI_SYSTEM":                  m.Sys.Name,                           // use CI_SYSTEM_NAME
		"CI_BRANCH":                  m.Curr.Commit.Branch,                 // use CI_COMMIT_BRANCH
		"CI_SOURCE_BRANCH":           sourceBranch,                         // use CI_COMMIT_SOURCE_BRANCH
		"CI_TARGET_BRANCH":           targetBranch,                         // use CI_COMMIT_TARGET_BRANCH
		"CI_TAG":                     "",                                   // use CI_COMMIT_TAG
		"CI_PULL_REQUEST":            "",                                   // use CI_COMMIT_PULL_REQUEST
	}
	if m.Curr.Event == EventTag {
		params["CI_COMMIT_TAG"] = strings.TrimPrefix(m.Curr.Commit.Ref, "refs/tags/")
		params["CI_TAG"] = params["CI_COMMIT_TAG"]
	}
	if m.Curr.Event == EventPull {
		params["CI_COMMIT_PULL_REQUEST"] = pullRegexp.FindString(m.Curr.Commit.Ref)
		params["CI_PULL_REQUEST"] = params["CI_COMMIT_PULL_REQUEST"]
	}

	return params
}

var pullRegexp = regexp.MustCompile(`\d+`)

func (m *Metadata) SetPlatform(platform string) {
	m.Sys.Arch = platform
}
