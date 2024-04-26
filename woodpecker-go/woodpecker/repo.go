package woodpecker

import "fmt"

const (
	pathRepoPost       = "%s/api/repos?forge_remote_id=%d"
	pathRepo           = "%s/api/repos/%d"
	pathRepoLookup     = "%s/api/repos/lookup/%s"
	pathRepoMove       = "%s/api/repos/%d/move?to=%s"
	pathChown          = "%s/api/repos/%d/chown"
	pathRepair         = "%s/api/repos/%d/repair"
	pathPipelines      = "%s/api/repos/%d/pipelines"
	pathPipeline       = "%s/api/repos/%d/pipelines/%v"
	pathPipelineLogs   = "%s/api/repos/%d/logs/%d"
	pathStepLogs       = "%s/api/repos/%d/logs/%d/%d"
	pathApprove        = "%s/api/repos/%d/pipelines/%d/approve"
	pathDecline        = "%s/api/repos/%d/pipelines/%d/decline"
	pathStop           = "%s/api/repos/%d/pipelines/%d/cancel"
	pathRepoSecrets    = "%s/api/repos/%d/secrets"
	pathRepoSecret     = "%s/api/repos/%d/secrets/%s"
	pathRepoRegistries = "%s/api/repos/%d/registry"
	pathRepoRegistry   = "%s/api/repos/%d/registry/%s"
	pathRepoCrons      = "%s/api/repos/%d/cron"
	pathRepoCron       = "%s/api/repos/%d/cron/%d"
)

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

// RepoPost activates a repository.
func (c *client) RepoPost(forgeRemoteID int64) (*Repo, error) {
	out := new(Repo)
	uri := fmt.Sprintf(pathRepoPost, c.addr, forgeRemoteID)
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

// RepoMove moves a repository.
func (c *client) RepoMove(repoID int64, newFullName string) error {
	uri := fmt.Sprintf(pathRepoMove, c.addr, repoID, newFullName)
	return c.post(uri, nil, nil)
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

// CronList returns a list of cronjobs for the specified repository.
func (c *client) CronList(repoID int64) ([]*Cron, error) {
	out := make([]*Cron, 0, 5)
	uri := fmt.Sprintf(pathRepoCrons, c.addr, repoID)
	return out, c.get(uri, &out)
}

// CronCreate creates a new cron job for the specified repository.
func (c *client) CronCreate(repoID int64, in *Cron) (*Cron, error) {
	out := new(Cron)
	uri := fmt.Sprintf(pathRepoCrons, c.addr, repoID)
	return out, c.post(uri, in, out)
}

// CronUpdate updates an existing cron job for the specified repository.
func (c *client) CronUpdate(repoID int64, in *Cron) (*Cron, error) {
	out := new(Cron)
	uri := fmt.Sprintf(pathRepoCron, c.addr, repoID, in.ID)
	err := c.patch(uri, in, out)
	return out, err
}

// CronDelete deletes a cron job by cron-id for the specified repository.
func (c *client) CronDelete(repoID, cronID int64) error {
	uri := fmt.Sprintf(pathRepoCron, c.addr, repoID, cronID)
	return c.delete(uri)
}

// CronGet returns a cron job by cron-id for the specified repository.
func (c *client) CronGet(repoID, cronID int64) (*Cron, error) {
	out := new(Cron)
	uri := fmt.Sprintf(pathRepoCron, c.addr, repoID, cronID)
	return out, c.get(uri, out)
}

// Pipeline returns a repository pipeline by pipeline-id.
func (c *client) Pipeline(repoID, pipeline int64) (*Pipeline, error) {
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

// PipelineCreate creates a new pipeline for the specified repository.
func (c *client) PipelineCreate(repoID int64, options *PipelineOptions) (*Pipeline, error) {
	var out *Pipeline
	uri := fmt.Sprintf(pathPipelines, c.addr, repoID)
	err := c.post(uri, options, &out)
	return out, err
}

// PipelineStart re-starts a stopped pipeline.
func (c *client) PipelineStart(repoID, pipeline int64, params map[string]string) (*Pipeline, error) {
	out := new(Pipeline)
	val := mapValues(params)
	uri := fmt.Sprintf(pathPipeline, c.addr, repoID, pipeline)
	err := c.post(uri+"?"+val.Encode(), nil, out)
	return out, err
}

// PipelineStop cancels the running step.
func (c *client) PipelineStop(repoID, pipeline int64) error {
	uri := fmt.Sprintf(pathStop, c.addr, repoID, pipeline)
	err := c.post(uri, nil, nil)
	return err
}

// PipelineApprove approves a blocked pipeline.
func (c *client) PipelineApprove(repoID, pipeline int64) (*Pipeline, error) {
	out := new(Pipeline)
	uri := fmt.Sprintf(pathApprove, c.addr, repoID, pipeline)
	err := c.post(uri, nil, out)
	return out, err
}

// PipelineDecline declines a blocked pipeline.
func (c *client) PipelineDecline(repoID, pipeline int64) (*Pipeline, error) {
	out := new(Pipeline)
	uri := fmt.Sprintf(pathDecline, c.addr, repoID, pipeline)
	err := c.post(uri, nil, out)
	return out, err
}

// PipelineKill force kills the running pipeline.
func (c *client) PipelineKill(repoID, pipeline int64) error {
	uri := fmt.Sprintf(pathPipeline, c.addr, repoID, pipeline)
	err := c.delete(uri)
	return err
}

// LogsPurge purges the pipeline all steps logs for the specified pipeline.
func (c *client) LogsPurge(repoID, pipeline int64) error {
	uri := fmt.Sprintf(pathPipelineLogs, c.addr, repoID, pipeline)
	err := c.delete(uri)
	return err
}

// Deploy triggers a deployment for an existing pipeline using the
// specified target environment.
func (c *client) Deploy(repoID, pipeline int64, env string, params map[string]string) (*Pipeline, error) {
	out := new(Pipeline)
	val := mapValues(params)
	val.Set("event", EventDeploy)
	val.Set("deploy_to", env)
	uri := fmt.Sprintf(pathPipeline, c.addr, repoID, pipeline)
	err := c.post(uri+"?"+val.Encode(), nil, out)
	return out, err
}

// StepLogEntries returns the pipeline logs for the specified step.
func (c *client) StepLogEntries(repoID, num, step int64) ([]*LogEntry, error) {
	uri := fmt.Sprintf(pathStepLogs, c.addr, repoID, num, step)
	var out []*LogEntry
	err := c.get(uri, &out)
	return out, err
}

// StepLogsPurge purges the pipeline logs for the specified step.
func (c *client) StepLogsPurge(repoID, pipelineNumber, stepID int64) error {
	uri := fmt.Sprintf(pathStepLogs, c.addr, repoID, pipelineNumber, stepID)
	err := c.delete(uri)
	return err
}
