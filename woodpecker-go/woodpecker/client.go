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
	pathRepo           = "%s/api/repos/%s/%s"
	pathRepoMove       = "%s/api/repos/%s/%s/move?to=%s"
	pathChown          = "%s/api/repos/%s/%s/chown"
	pathRepair         = "%s/api/repos/%s/%s/repair"
	pathPipelines      = "%s/api/repos/%s/%s/pipelines"
	pathPipeline       = "%s/api/repos/%s/%s/pipelines/%v"
	pathLogs           = "%s/api/repos/%s/%s/logs/%d/%d"
	pathApprove        = "%s/api/repos/%s/%s/pipelines/%d/approve"
	pathDecline        = "%s/api/repos/%s/%s/pipelines/%d/decline"
	pathJob            = "%s/api/repos/%s/%s/pipelines/%d/%d"
	pathLogPurge       = "%s/api/repos/%s/%s/logs/%d"
	pathRepoSecrets    = "%s/api/repos/%s/%s/secrets"
	pathRepoSecret     = "%s/api/repos/%s/%s/secrets/%s"
	pathRepoRegistries = "%s/api/repos/%s/%s/registry"
	pathRepoRegistry   = "%s/api/repos/%s/%s/registry/%s"
	pathRepoCrons      = "%s/api/repos/%s/%s/cron"
	pathRepoCron       = "%s/api/repos/%s/%s/cron/%d"
	pathOrgSecrets     = "%s/api/orgs/%s/secrets"
	pathOrgSecret      = "%s/api/orgs/%s/secrets/%s"
	pathGlobalSecrets  = "%s/api/secrets"
	pathGlobalSecret   = "%s/api/secrets/%s"
	pathUsers          = "%s/api/users"
	pathUser           = "%s/api/users/%s"
	pathPipelineQueue  = "%s/api/pipelines"
	pathQueue          = "%s/api/queue"
	pathLogLevel       = "%s/api/log-level"
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

// Repo returns a repository by name.
func (c *client) Repo(owner, name string) (*Repo, error) {
	out := new(Repo)
	uri := fmt.Sprintf(pathRepo, c.addr, owner, name)
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
func (c *client) RepoListOpts(sync, all bool) ([]*Repo, error) {
	var out []*Repo
	uri := fmt.Sprintf(pathRepos+"?flush=%v&all=%v", c.addr, sync, all)
	err := c.get(uri, &out)
	return out, err
}

// RepoPost activates a repository.
func (c *client) RepoPost(owner, name string) (*Repo, error) {
	out := new(Repo)
	uri := fmt.Sprintf(pathRepo, c.addr, owner, name)
	err := c.post(uri, nil, out)
	return out, err
}

// RepoChown updates a repository owner.
func (c *client) RepoChown(owner, name string) (*Repo, error) {
	out := new(Repo)
	uri := fmt.Sprintf(pathChown, c.addr, owner, name)
	err := c.post(uri, nil, out)
	return out, err
}

// RepoRepair repairs the repository hooks.
func (c *client) RepoRepair(owner, name string) error {
	uri := fmt.Sprintf(pathRepair, c.addr, owner, name)
	return c.post(uri, nil, nil)
}

// RepoPatch updates a repository.
func (c *client) RepoPatch(owner, name string, in *RepoPatch) (*Repo, error) {
	out := new(Repo)
	uri := fmt.Sprintf(pathRepo, c.addr, owner, name)
	err := c.patch(uri, in, out)
	return out, err
}

// RepoDel deletes a repository.
func (c *client) RepoDel(owner, name string) error {
	uri := fmt.Sprintf(pathRepo, c.addr, owner, name)
	err := c.delete(uri)
	return err
}

// RepoMove moves a repository
func (c *client) RepoMove(owner, name, newFullName string) error {
	uri := fmt.Sprintf(pathRepoMove, c.addr, owner, name, newFullName)
	return c.post(uri, nil, nil)
}

// Pipeline returns a repository pipeline by number.
func (c *client) Pipeline(owner, name string, num int) (*Pipeline, error) {
	out := new(Pipeline)
	uri := fmt.Sprintf(pathPipeline, c.addr, owner, name, num)
	err := c.get(uri, out)
	return out, err
}

// Pipeline returns the latest repository pipeline by branch.
func (c *client) PipelineLast(owner, name, branch string) (*Pipeline, error) {
	out := new(Pipeline)
	uri := fmt.Sprintf(pathPipeline, c.addr, owner, name, "latest")
	if len(branch) != 0 {
		uri += "?branch=" + branch
	}
	err := c.get(uri, out)
	return out, err
}

// PipelineList returns a list of recent pipelines for the
// the specified repository.
func (c *client) PipelineList(owner, name string) ([]*Pipeline, error) {
	var out []*Pipeline
	uri := fmt.Sprintf(pathPipelines, c.addr, owner, name)
	err := c.get(uri, &out)
	return out, err
}

func (c *client) PipelineCreate(owner, name string, options *PipelineOptions) (*Pipeline, error) {
	var out *Pipeline
	uri := fmt.Sprintf(pathPipelines, c.addr, owner, name)
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
func (c *client) PipelineStart(owner, name string, num int, params map[string]string) (*Pipeline, error) {
	out := new(Pipeline)
	val := mapValues(params)
	uri := fmt.Sprintf(pathPipeline, c.addr, owner, name, num)
	err := c.post(uri+"?"+val.Encode(), nil, out)
	return out, err
}

// PipelineStop cancels the running job.
func (c *client) PipelineStop(owner, name string, num, job int) error {
	uri := fmt.Sprintf(pathJob, c.addr, owner, name, num, job)
	err := c.delete(uri)
	return err
}

// PipelineApprove approves a blocked pipeline.
func (c *client) PipelineApprove(owner, name string, num int) (*Pipeline, error) {
	out := new(Pipeline)
	uri := fmt.Sprintf(pathApprove, c.addr, owner, name, num)
	err := c.post(uri, nil, out)
	return out, err
}

// PipelineDecline declines a blocked pipeline.
func (c *client) PipelineDecline(owner, name string, num int) (*Pipeline, error) {
	out := new(Pipeline)
	uri := fmt.Sprintf(pathDecline, c.addr, owner, name, num)
	err := c.post(uri, nil, out)
	return out, err
}

// PipelineKill force kills the running pipeline.
func (c *client) PipelineKill(owner, name string, num int) error {
	uri := fmt.Sprintf(pathPipeline, c.addr, owner, name, num)
	err := c.delete(uri)
	return err
}

// PipelineLogs returns the pipeline logs for the specified job.
func (c *client) PipelineLogs(owner, name string, num, job int) ([]*Logs, error) {
	uri := fmt.Sprintf(pathLogs, c.addr, owner, name, num, job)
	var out []*Logs
	err := c.get(uri, &out)
	return out, err
}

// Deploy triggers a deployment for an existing pipeline using the
// specified target environment.
func (c *client) Deploy(owner, name string, num int, env string, params map[string]string) (*Pipeline, error) {
	out := new(Pipeline)
	val := mapValues(params)
	val.Set("event", "deployment")
	val.Set("deploy_to", env)
	uri := fmt.Sprintf(pathPipeline, c.addr, owner, name, num)
	err := c.post(uri+"?"+val.Encode(), nil, out)
	return out, err
}

// LogsPurge purges the pipeline logs for the specified pipeline.
func (c *client) LogsPurge(owner, name string, num int) error {
	uri := fmt.Sprintf(pathLogPurge, c.addr, owner, name, num)
	err := c.delete(uri)
	return err
}

// Registry returns a registry by hostname.
func (c *client) Registry(owner, name, hostname string) (*Registry, error) {
	out := new(Registry)
	uri := fmt.Sprintf(pathRepoRegistry, c.addr, owner, name, hostname)
	err := c.get(uri, out)
	return out, err
}

// RegistryList returns a list of all repository registries.
func (c *client) RegistryList(owner, name string) ([]*Registry, error) {
	var out []*Registry
	uri := fmt.Sprintf(pathRepoRegistries, c.addr, owner, name)
	err := c.get(uri, &out)
	return out, err
}

// RegistryCreate creates a registry.
func (c *client) RegistryCreate(owner, name string, in *Registry) (*Registry, error) {
	out := new(Registry)
	uri := fmt.Sprintf(pathRepoRegistries, c.addr, owner, name)
	err := c.post(uri, in, out)
	return out, err
}

// RegistryUpdate updates a registry.
func (c *client) RegistryUpdate(owner, name string, in *Registry) (*Registry, error) {
	out := new(Registry)
	uri := fmt.Sprintf(pathRepoRegistry, c.addr, owner, name, in.Address)
	err := c.patch(uri, in, out)
	return out, err
}

// RegistryDelete deletes a registry.
func (c *client) RegistryDelete(owner, name, hostname string) error {
	uri := fmt.Sprintf(pathRepoRegistry, c.addr, owner, name, hostname)
	return c.delete(uri)
}

// Secret returns a secret by name.
func (c *client) Secret(owner, name, secret string) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathRepoSecret, c.addr, owner, name, secret)
	err := c.get(uri, out)
	return out, err
}

// SecretList returns a list of all repository secrets.
func (c *client) SecretList(owner, name string) ([]*Secret, error) {
	var out []*Secret
	uri := fmt.Sprintf(pathRepoSecrets, c.addr, owner, name)
	err := c.get(uri, &out)
	return out, err
}

// SecretCreate creates a secret.
func (c *client) SecretCreate(owner, name string, in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathRepoSecrets, c.addr, owner, name)
	err := c.post(uri, in, out)
	return out, err
}

// SecretUpdate updates a secret.
func (c *client) SecretUpdate(owner, name string, in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathRepoSecret, c.addr, owner, name, in.Name)
	err := c.patch(uri, in, out)
	return out, err
}

// SecretDelete deletes a secret.
func (c *client) SecretDelete(owner, name, secret string) error {
	uri := fmt.Sprintf(pathRepoSecret, c.addr, owner, name, secret)
	return c.delete(uri)
}

// OrgSecret returns an organization secret by name.
func (c *client) OrgSecret(owner, secret string) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathOrgSecret, c.addr, owner, secret)
	err := c.get(uri, out)
	return out, err
}

// OrgSecretList returns a list of all organization secrets.
func (c *client) OrgSecretList(owner string) ([]*Secret, error) {
	var out []*Secret
	uri := fmt.Sprintf(pathOrgSecrets, c.addr, owner)
	err := c.get(uri, &out)
	return out, err
}

// OrgSecretCreate creates an organization secret.
func (c *client) OrgSecretCreate(owner string, in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathOrgSecrets, c.addr, owner)
	err := c.post(uri, in, out)
	return out, err
}

// OrgSecretUpdate updates an organization secret.
func (c *client) OrgSecretUpdate(owner string, in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathOrgSecret, c.addr, owner, in.Name)
	err := c.patch(uri, in, out)
	return out, err
}

// OrgSecretDelete deletes an organization secret.
func (c *client) OrgSecretDelete(owner, secret string) error {
	uri := fmt.Sprintf(pathOrgSecret, c.addr, owner, secret)
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

func (c *client) CronList(owner, repo string) ([]*Cron, error) {
	out := make([]*Cron, 0, 5)
	uri := fmt.Sprintf(pathRepoCrons, c.addr, owner, repo)
	return out, c.get(uri, &out)
}

func (c *client) CronCreate(owner, repo string, in *Cron) (*Cron, error) {
	out := new(Cron)
	uri := fmt.Sprintf(pathRepoCrons, c.addr, owner, repo)
	return out, c.post(uri, in, out)
}

func (c *client) CronUpdate(owner, repo string, in *Cron) (*Cron, error) {
	out := new(Cron)
	uri := fmt.Sprintf(pathRepoCron, c.addr, owner, repo, in.ID)
	err := c.patch(uri, in, out)
	return out, err
}

func (c *client) CronDelete(owner, repo string, cronID int64) error {
	uri := fmt.Sprintf(pathRepoCron, c.addr, owner, repo, cronID)
	return c.delete(uri)
}

func (c *client) CronGet(owner, repo string, cronID int64) (*Cron, error) {
	out := new(Cron)
	uri := fmt.Sprintf(pathRepoCron, c.addr, owner, repo, cronID)
	return out, c.get(uri, out)
}

//
// http request helper functions
//

// helper function for making an http GET request.
func (c *client) get(rawurl string, out interface{}) error {
	return c.do(rawurl, "GET", nil, out)
}

// helper function for making an http POST request.
func (c *client) post(rawurl string, in, out interface{}) error {
	return c.do(rawurl, "POST", in, out)
}

// helper function for making an http PUT request.
func (c *client) put(rawurl string, in, out interface{}) error {
	return c.do(rawurl, "PUT", in, out)
}

// helper function for making an http PATCH request.
func (c *client) patch(rawurl string, in, out interface{}) error {
	return c.do(rawurl, "PATCH", in, out)
}

// helper function for making an http DELETE request.
func (c *client) delete(rawurl string) error {
	return c.do(rawurl, "DELETE", nil, nil)
}

// helper function to make an http request
func (c *client) do(rawurl, method string, in, out interface{}) error {
	body, err := c.open(rawurl, method, in, out)
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
func (c *client) open(rawurl, method string, in, out interface{}) (io.ReadCloser, error) {
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
