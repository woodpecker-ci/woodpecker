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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	pathSelf           = "%s/api/user"
	pathRepos          = "%s/api/user/repos"
	pathRepoPost       = "%s/api/repos"
	pathRepo           = "%s/api/repos/%d"
	pathRepoLookup     = "%s/api/repos/lookup/%s"
	pathRepoMove       = "%s/api/repos/%d/move?to=%s"
	pathChown          = "%s/api/repos/%d/chown"
	pathRepair         = "%s/api/repos/%d/repair"
	pathPipelines      = "%s/api/repos/%d/pipelines"
	pathPipeline       = "%s/api/repos/%d/pipelines/%v"
	pathLogs           = "%s/api/repos/%d/logs/%d/%d"
	pathApprove        = "%s/api/repos/%d/pipelines/%d/approve"
	pathDecline        = "%s/api/repos/%d/pipelines/%d/decline"
	pathStop           = "%s/api/repos/%d/pipelines/%d/cancel"
	pathLogPurge       = "%s/api/repos/%d/logs/%d"
	pathRepoSecrets    = "%s/api/repos/%d/secrets"
	pathRepoSecret     = "%s/api/repos/%d/secrets/%s"
	pathRepoRegistries = "%s/api/repos/%d/registry"
	pathRepoRegistry   = "%s/api/repos/%d/registry/%s"
	pathRepoCrons      = "%s/api/repos/%d/cron"
	pathRepoCron       = "%s/api/repos/%d/cron/%d"
	pathOrg            = "%s/api/orgs/%d"
	pathOrgLookup      = "%s/api/orgs/lookup/%s"
	pathOrgSecrets     = "%s/api/orgs/%d/secrets"
	pathOrgSecret      = "%s/api/orgs/%d/secrets/%s"
	pathGlobalSecrets  = "%s/api/secrets"
	pathGlobalSecret   = "%s/api/secrets/%s"
	pathUsers          = "%s/api/users"
	pathUser           = "%s/api/users/%s"
	pathPipelineQueue  = "%s/api/pipelines"
	pathQueue          = "%s/api/queue"
	pathLogLevel       = "%s/api/log-level"
	pathAgents         = "%s/api/agents"
	pathAgent          = "%s/api/agents/%d"
	pathAgentTasks     = "%s/api/agents/%d/tasks"
	// TODO: implement endpoints
	// pathFeed           = "%s/api/user/feed"
	// pathVersion        = "%s/version"
)

type client struct {
	client *http.Client
	addr   string
}

// New returns a client at the specified url.
func New(uri string) Client {
	return &client{http.DefaultClient, strings.TrimSuffix(uri, "/")}
}

// NewClient returns a client at the specified url.
func NewClient(uri string, cli *http.Client) Client {
	return &client{cli, strings.TrimSuffix(uri, "/")}
}

// SetClient sets the http.Client.
func (c *client) SetClient(client *http.Client) {
	c.client = client
}

// SetAddress sets the server address.
func (c *client) SetAddress(addr string) {
	c.addr = addr
}

// Self returns the currently authenticated user.
func (c *client) Self() (*User, error) {
	out := new(User)
	uri := fmt.Sprintf(pathSelf, c.addr)
	err := c.get(uri, out)
	return out, err
}

// User returns a user by login.
func (c *client) User(login string) (*User, error) {
	out := new(User)
	uri := fmt.Sprintf(pathUser, c.addr, login)
	err := c.get(uri, out)
	return out, err
}

// UserList returns a list of all registered users.
func (c *client) UserList() ([]*User, error) {
	var out []*User
	uri := fmt.Sprintf(pathUsers, c.addr)
	err := c.get(uri, &out)
	return out, err
}

// UserPost creates a new user account.
func (c *client) UserPost(in *User) (*User, error) {
	out := new(User)
	uri := fmt.Sprintf(pathUsers, c.addr)
	err := c.post(uri, in, out)
	return out, err
}

// UserPatch updates a user account.
func (c *client) UserPatch(in *User) (*User, error) {
	out := new(User)
	uri := fmt.Sprintf(pathUser, c.addr, in.Login)
	err := c.patch(uri, in, out)
	return out, err
}

// UserDel deletes a user account.
func (c *client) UserDel(login string) error {
	uri := fmt.Sprintf(pathUser, c.addr, login)
	err := c.delete(uri)
	return err
}

// Repo returns a repository by id.
func (c *client) Repo(repoID int64) (*Repo, error) {
	out := new(Repo)
	uri := fmt.Sprintf(pathRepo, c.addr, repoID)
	err := c.get(uri, out)
	return out, err
}

// RepoLookup returns a repository by name.
func (c *client) RepoLookup(fullName string) (*Repo, error) {
	out := new(Repo)
	uri := fmt.Sprintf(pathRepoLookup, c.addr, fullName)
	err := c.get(uri, out)
	return out, err
}

// RepoList returns a list of all repositories to which
// the user has explicit access in the host system.
func (c *client) RepoList() ([]*Repo, error) {
	var out []*Repo
	uri := fmt.Sprintf(pathRepos, c.addr)
	err := c.get(uri, &out)
	return out, err
}

// RepoListOpts returns a list of all repositories to which
// the user has explicit access in the host system.
func (c *client) RepoListOpts(all bool) ([]*Repo, error) {
	var out []*Repo
	uri := fmt.Sprintf(pathRepos+"?all=%v", c.addr, all)
	err := c.get(uri, &out)
	return out, err
}

// RepoPost activates a repository.
func (c *client) RepoPost(forgeRemoteID int64) (*Repo, error) {
	out := new(Repo)
	uri := fmt.Sprintf(pathRepoPost+"?forge_remote_id=%d", c.addr, forgeRemoteID)
	err := c.post(uri, nil, out)
	return out, err
}

// RepoChown updates a repository owner.
func (c *client) RepoChown(repoID int64) (*Repo, error) {
	out := new(Repo)
	uri := fmt.Sprintf(pathChown, c.addr, repoID)
	err := c.post(uri, nil, out)
	return out, err
}

// RepoRepair repairs the repository hooks.
func (c *client) RepoRepair(repoID int64) error {
	uri := fmt.Sprintf(pathRepair, c.addr, repoID)
	return c.post(uri, nil, nil)
}

// RepoPatch updates a repository.
func (c *client) RepoPatch(repoID int64, in *RepoPatch) (*Repo, error) {
	out := new(Repo)
	uri := fmt.Sprintf(pathRepo, c.addr, repoID)
	err := c.patch(uri, in, out)
	return out, err
}

// RepoDel deletes a repository.
func (c *client) RepoDel(repoID int64) error {
	uri := fmt.Sprintf(pathRepo, c.addr, repoID)
	err := c.delete(uri)
	return err
}

// RepoMove moves a repository
func (c *client) RepoMove(repoID int64, newFullName string) error {
	uri := fmt.Sprintf(pathRepoMove, c.addr, repoID, newFullName)
	return c.post(uri, nil, nil)
}

// Pipeline returns a repository pipeline by pipeline-id.
func (c *client) Pipeline(repoID int64, pipeline int) (*Pipeline, error) {
	out := new(Pipeline)
	uri := fmt.Sprintf(pathPipeline, c.addr, repoID, pipeline)
	err := c.get(uri, out)
	return out, err
}

// Pipeline returns the latest repository pipeline by branch.
func (c *client) PipelineLast(repoID int64, branch string) (*Pipeline, error) {
	out := new(Pipeline)
	uri := fmt.Sprintf(pathPipeline, c.addr, repoID, "latest")
	if len(branch) != 0 {
		uri += "?branch=" + branch
	}
	err := c.get(uri, out)
	return out, err
}

// PipelineList returns a list of recent pipelines for the
// the specified repository.
func (c *client) PipelineList(repoID int64) ([]*Pipeline, error) {
	var out []*Pipeline
	uri := fmt.Sprintf(pathPipelines, c.addr, repoID)
	err := c.get(uri, &out)
	return out, err
}

func (c *client) PipelineCreate(repoID int64, options *PipelineOptions) (*Pipeline, error) {
	var out *Pipeline
	uri := fmt.Sprintf(pathPipelines, c.addr, repoID)
	err := c.post(uri, options, &out)
	return out, err
}

// PipelineQueue returns a list of enqueued pipelines.
func (c *client) PipelineQueue() ([]*Activity, error) {
	var out []*Activity
	uri := fmt.Sprintf(pathPipelineQueue, c.addr)
	err := c.get(uri, &out)
	return out, err
}

// PipelineStart re-starts a stopped pipeline.
func (c *client) PipelineStart(repoID int64, pipeline int, params map[string]string) (*Pipeline, error) {
	out := new(Pipeline)
	val := mapValues(params)
	uri := fmt.Sprintf(pathPipeline, c.addr, repoID, pipeline)
	err := c.post(uri+"?"+val.Encode(), nil, out)
	return out, err
}

// PipelineStop cancels the running step.
func (c *client) PipelineStop(repoID int64, pipeline int) error {
	uri := fmt.Sprintf(pathStop, c.addr, repoID, pipeline)
	err := c.post(uri, nil, nil)
	return err
}

// PipelineApprove approves a blocked pipeline.
func (c *client) PipelineApprove(repoID int64, pipeline int) (*Pipeline, error) {
	out := new(Pipeline)
	uri := fmt.Sprintf(pathApprove, c.addr, repoID, pipeline)
	err := c.post(uri, nil, out)
	return out, err
}

// PipelineDecline declines a blocked pipeline.
func (c *client) PipelineDecline(repoID int64, pipeline int) (*Pipeline, error) {
	out := new(Pipeline)
	uri := fmt.Sprintf(pathDecline, c.addr, repoID, pipeline)
	err := c.post(uri, nil, out)
	return out, err
}

// PipelineKill force kills the running pipeline.
func (c *client) PipelineKill(repoID int64, pipeline int) error {
	uri := fmt.Sprintf(pathPipeline, c.addr, repoID, pipeline)
	err := c.delete(uri)
	return err
}

// PipelineLogs returns the pipeline logs for the specified step.
func (c *client) StepLogEntries(repoID int64, num, step int) ([]*LogEntry, error) {
	uri := fmt.Sprintf(pathLogs, c.addr, repoID, num, step)
	var out []*LogEntry
	err := c.get(uri, &out)
	return out, err
}

// Deploy triggers a deployment for an existing pipeline using the
// specified target environment.
func (c *client) Deploy(repoID int64, pipeline int, env string, params map[string]string) (*Pipeline, error) {
	out := new(Pipeline)
	val := mapValues(params)
	val.Set("event", EventDeploy)
	val.Set("deploy_to", env)
	uri := fmt.Sprintf(pathPipeline, c.addr, repoID, pipeline)
	err := c.post(uri+"?"+val.Encode(), nil, out)
	return out, err
}

// LogsPurge purges the pipeline logs for the specified pipeline.
func (c *client) LogsPurge(repoID int64, pipeline int) error {
	uri := fmt.Sprintf(pathLogPurge, c.addr, repoID, pipeline)
	err := c.delete(uri)
	return err
}

// Registry returns a registry by hostname.
func (c *client) Registry(repoID int64, hostname string) (*Registry, error) {
	out := new(Registry)
	uri := fmt.Sprintf(pathRepoRegistry, c.addr, repoID, hostname)
	err := c.get(uri, out)
	return out, err
}

// RegistryList returns a list of all repository registries.
func (c *client) RegistryList(repoID int64) ([]*Registry, error) {
	var out []*Registry
	uri := fmt.Sprintf(pathRepoRegistries, c.addr, repoID)
	err := c.get(uri, &out)
	return out, err
}

// RegistryCreate creates a registry.
func (c *client) RegistryCreate(repoID int64, in *Registry) (*Registry, error) {
	out := new(Registry)
	uri := fmt.Sprintf(pathRepoRegistries, c.addr, repoID)
	err := c.post(uri, in, out)
	return out, err
}

// RegistryUpdate updates a registry.
func (c *client) RegistryUpdate(repoID int64, in *Registry) (*Registry, error) {
	out := new(Registry)
	uri := fmt.Sprintf(pathRepoRegistry, c.addr, repoID, in.Address)
	err := c.patch(uri, in, out)
	return out, err
}

// RegistryDelete deletes a registry.
func (c *client) RegistryDelete(repoID int64, hostname string) error {
	uri := fmt.Sprintf(pathRepoRegistry, c.addr, repoID, hostname)
	return c.delete(uri)
}

// Secret returns a secret by name.
func (c *client) Secret(repoID int64, secret string) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathRepoSecret, c.addr, repoID, secret)
	err := c.get(uri, out)
	return out, err
}

// SecretList returns a list of all repository secrets.
func (c *client) SecretList(repoID int64) ([]*Secret, error) {
	var out []*Secret
	uri := fmt.Sprintf(pathRepoSecrets, c.addr, repoID)
	err := c.get(uri, &out)
	return out, err
}

// SecretCreate creates a secret.
func (c *client) SecretCreate(repoID int64, in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathRepoSecrets, c.addr, repoID)
	err := c.post(uri, in, out)
	return out, err
}

// SecretUpdate updates a secret.
func (c *client) SecretUpdate(repoID int64, in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathRepoSecret, c.addr, repoID, in.Name)
	err := c.patch(uri, in, out)
	return out, err
}

// SecretDelete deletes a secret.
func (c *client) SecretDelete(repoID int64, secret string) error {
	uri := fmt.Sprintf(pathRepoSecret, c.addr, repoID, secret)
	return c.delete(uri)
}

// Org returns an organization by id.
func (c *client) Org(orgID int64) (*Org, error) {
	out := new(Org)
	uri := fmt.Sprintf(pathOrg, c.addr, orgID)
	err := c.get(uri, out)
	return out, err
}

// OrgLookup returns a organsization by its name.
func (c *client) OrgLookup(name string) (*Org, error) {
	out := new(Org)
	uri := fmt.Sprintf(pathOrgLookup, c.addr, name)
	err := c.get(uri, out)
	return out, err
}

// OrgSecret returns an organization secret by name.
func (c *client) OrgSecret(orgID int64, secret string) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathOrgSecret, c.addr, orgID, secret)
	err := c.get(uri, out)
	return out, err
}

// OrgSecretList returns a list of all organization secrets.
func (c *client) OrgSecretList(orgID int64) ([]*Secret, error) {
	var out []*Secret
	uri := fmt.Sprintf(pathOrgSecrets, c.addr, orgID)
	err := c.get(uri, &out)
	return out, err
}

// OrgSecretCreate creates an organization secret.
func (c *client) OrgSecretCreate(orgID int64, in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathOrgSecrets, c.addr, orgID)
	err := c.post(uri, in, out)
	return out, err
}

// OrgSecretUpdate updates an organization secret.
func (c *client) OrgSecretUpdate(orgID int64, in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathOrgSecret, c.addr, orgID, in.Name)
	err := c.patch(uri, in, out)
	return out, err
}

// OrgSecretDelete deletes an organization secret.
func (c *client) OrgSecretDelete(orgID int64, secret string) error {
	uri := fmt.Sprintf(pathOrgSecret, c.addr, orgID, secret)
	return c.delete(uri)
}

// GlobalOrgSecret returns an global secret by name.
func (c *client) GlobalSecret(secret string) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathGlobalSecret, c.addr, secret)
	err := c.get(uri, out)
	return out, err
}

// GlobalSecretList returns a list of all global secrets.
func (c *client) GlobalSecretList() ([]*Secret, error) {
	var out []*Secret
	uri := fmt.Sprintf(pathGlobalSecrets, c.addr)
	err := c.get(uri, &out)
	return out, err
}

// GlobalSecretCreate creates a global secret.
func (c *client) GlobalSecretCreate(in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathGlobalSecrets, c.addr)
	err := c.post(uri, in, out)
	return out, err
}

// GlobalSecretUpdate updates a global secret.
func (c *client) GlobalSecretUpdate(in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathGlobalSecret, c.addr, in.Name)
	err := c.patch(uri, in, out)
	return out, err
}

// GlobalSecretDelete deletes a global secret.
func (c *client) GlobalSecretDelete(secret string) error {
	uri := fmt.Sprintf(pathGlobalSecret, c.addr, secret)
	return c.delete(uri)
}

// QueueInfo returns queue info
func (c *client) QueueInfo() (*Info, error) {
	out := new(Info)
	uri := fmt.Sprintf(pathQueue+"/info", c.addr)
	err := c.get(uri, out)
	return out, err
}

// LogLevel returns the current logging level
func (c *client) LogLevel() (*LogLevel, error) {
	out := new(LogLevel)
	uri := fmt.Sprintf(pathLogLevel, c.addr)
	err := c.get(uri, out)
	return out, err
}

// SetLogLevel sets the logging level of the server
func (c *client) SetLogLevel(in *LogLevel) (*LogLevel, error) {
	out := new(LogLevel)
	uri := fmt.Sprintf(pathLogLevel, c.addr)
	err := c.post(uri, in, out)
	return out, err
}

func (c *client) CronList(repoID int64) ([]*Cron, error) {
	out := make([]*Cron, 0, 5)
	uri := fmt.Sprintf(pathRepoCrons, c.addr, repoID)
	return out, c.get(uri, &out)
}

func (c *client) CronCreate(repoID int64, in *Cron) (*Cron, error) {
	out := new(Cron)
	uri := fmt.Sprintf(pathRepoCrons, c.addr, repoID)
	return out, c.post(uri, in, out)
}

func (c *client) CronUpdate(repoID int64, in *Cron) (*Cron, error) {
	out := new(Cron)
	uri := fmt.Sprintf(pathRepoCron, c.addr, repoID, in.ID)
	err := c.patch(uri, in, out)
	return out, err
}

func (c *client) CronDelete(repoID, cronID int64) error {
	uri := fmt.Sprintf(pathRepoCron, c.addr, repoID, cronID)
	return c.delete(uri)
}

func (c *client) CronGet(repoID, cronID int64) (*Cron, error) {
	out := new(Cron)
	uri := fmt.Sprintf(pathRepoCron, c.addr, repoID, cronID)
	return out, c.get(uri, out)
}

func (c *client) AgentList() ([]*Agent, error) {
	out := make([]*Agent, 0, 5)
	uri := fmt.Sprintf(pathAgents, c.addr)
	return out, c.get(uri, &out)
}

func (c *client) Agent(agentID int64) (*Agent, error) {
	out := new(Agent)
	uri := fmt.Sprintf(pathAgent, c.addr, agentID)
	return out, c.get(uri, out)
}

func (c *client) AgentCreate(in *Agent) (*Agent, error) {
	out := new(Agent)
	uri := fmt.Sprintf(pathAgents, c.addr)
	return out, c.post(uri, in, out)
}

func (c *client) AgentUpdate(in *Agent) (*Agent, error) {
	out := new(Agent)
	uri := fmt.Sprintf(pathAgent, c.addr, in.ID)
	return out, c.patch(uri, in, out)
}

func (c *client) AgentDelete(agentID int64) error {
	uri := fmt.Sprintf(pathAgent, c.addr, agentID)
	return c.delete(uri)
}

func (c *client) AgentTasksList(agentID int64) ([]*Task, error) {
	out := make([]*Task, 0, 5)
	uri := fmt.Sprintf(pathAgentTasks, c.addr, agentID)
	return out, c.get(uri, &out)
}

//
// http request helper functions
//

// helper function for making an http GET request.
func (c *client) get(rawurl string, out interface{}) error {
	return c.do(rawurl, http.MethodGet, nil, out)
}

// helper function for making an http POST request.
func (c *client) post(rawurl string, in, out interface{}) error {
	return c.do(rawurl, http.MethodPost, in, out)
}

// helper function for making an http PATCH request.
func (c *client) patch(rawurl string, in, out interface{}) error {
	return c.do(rawurl, http.MethodPatch, in, out)
}

// helper function for making an http DELETE request.
func (c *client) delete(rawurl string) error {
	return c.do(rawurl, http.MethodDelete, nil, nil)
}

// helper function to make an http request
func (c *client) do(rawurl, method string, in, out interface{}) error {
	body, err := c.open(rawurl, method, in)
	if err != nil {
		return err
	}
	defer body.Close()
	if out != nil {
		return json.NewDecoder(body).Decode(out)
	}
	return nil
}

// helper function to open an http request
func (c *client) open(rawurl, method string, in interface{}) (io.ReadCloser, error) {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, uri.String(), nil)
	if err != nil {
		return nil, err
	}
	if in != nil {
		decoded, derr := json.Marshal(in)
		if derr != nil {
			return nil, derr
		}
		buf := bytes.NewBuffer(decoded)
		req.Body = io.NopCloser(buf)
		req.ContentLength = int64(len(decoded))
		req.Header.Set("Content-Length", strconv.Itoa(len(decoded)))
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > http.StatusPartialContent {
		defer resp.Body.Close()
		out, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("client error %d: %s", resp.StatusCode, string(out))
	}
	return resp.Body, nil
}

// mapValues converts a map to url.Values
func mapValues(params map[string]string) url.Values {
	values := url.Values{}
	for key, val := range params {
		values.Add(key, val)
	}
	return values
}
