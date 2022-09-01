package woodpecker

import (
	"net/http"
)

// Client is used to communicate with a Woodpecker server.
type Client interface {
	// SetClient sets the http.Client.
	SetClient(*http.Client)

	// SetAddress sets the server address.
	SetAddress(string)

	// Self returns the currently authenticated user.
	Self() (*User, error)

	// User returns a user by login.
	User(string) (*User, error)

	// UserList returns a list of all registered users.
	UserList() ([]*User, error)

	// UserPost creates a new user account.
	UserPost(*User) (*User, error)

	// UserPatch updates a user account.
	UserPatch(*User) (*User, error)

	// UserDel deletes a user account.
	UserDel(string) error

	// Repo returns a repository by name.
	Repo(string, string) (*Repo, error)

	// RepoList returns a list of all repositories to which the user has explicit
	// access in the host system.
	RepoList() ([]*Repo, error)

	// RepoListOpts returns a list of all repositories to which the user has
	// explicit access in the host system.
	RepoListOpts(bool, bool) ([]*Repo, error)

	// RepoPost activates a repository.
	RepoPost(string, string) (*Repo, error)

	// RepoPatch updates a repository.
	RepoPatch(string, string, *RepoPatch) (*Repo, error)

	// RepoMove moves the repository
	RepoMove(string, string, string) error

	// RepoChown updates a repository owner.
	RepoChown(string, string) (*Repo, error)

	// RepoRepair repairs the repository hooks.
	RepoRepair(string, string) error

	// RepoDel deletes a repository.
	RepoDel(string, string) error

	// Build returns a repository build by number.
	Build(string, string, int) (*Build, error)

	// BuildLast returns the latest repository build by branch. An empty branch
	// will result in the default branch.
	BuildLast(string, string, string) (*Build, error)

	// BuildList returns a list of recent builds for the
	// the specified repository.
	BuildList(string, string) ([]*Build, error)

	// BuildQueue returns a list of enqueued builds.
	BuildQueue() ([]*Activity, error)

	// BuildStart re-starts a stopped build.
	BuildStart(string, string, int, map[string]string) (*Build, error)

	// BuildStop stops the specified running job for given build.
	BuildStop(string, string, int, int) error

	// BuildApprove approves a blocked build.
	BuildApprove(string, string, int) (*Build, error)

	// BuildDecline declines a blocked build.
	BuildDecline(string, string, int) (*Build, error)

	// BuildKill force kills the running build.
	BuildKill(string, string, int) error

	// BuildLogs returns the logs for the given build
	BuildLogs(string, string, int, int) ([]*Logs, error)

	// Deploy triggers a deployment for an existing build using the specified
	// target environment.
	Deploy(string, string, int, string, map[string]string) (*Build, error)

	// LogsPurge purges the build logs for the specified build.
	LogsPurge(string, string, int) error

	// Registry returns a registry by hostname.
	Registry(owner, name, hostname string) (*Registry, error)

	// RegistryList returns a list of all repository registries.
	RegistryList(owner, name string) ([]*Registry, error)

	// RegistryCreate creates a registry.
	RegistryCreate(owner, name string, registry *Registry) (*Registry, error)

	// RegistryUpdate updates a registry.
	RegistryUpdate(owner, name string, registry *Registry) (*Registry, error)

	// RegistryDelete deletes a registry.
	RegistryDelete(owner, name, hostname string) error

	// Secret returns a secret by name.
	Secret(owner, name, secret string) (*Secret, error)

	// SecretList returns a list of all repository secrets.
	SecretList(owner, name string) ([]*Secret, error)

	// SecretCreate creates a secret.
	SecretCreate(owner, name string, secret *Secret) (*Secret, error)

	// SecretUpdate updates a secret.
	SecretUpdate(owner, name string, secret *Secret) (*Secret, error)

	// SecretDelete deletes a secret.
	SecretDelete(owner, name, secret string) error

	// OrgSecret returns an organization secret by name.
	OrgSecret(owner, secret string) (*Secret, error)

	// OrgSecretList returns a list of all organization secrets.
	OrgSecretList(owner string) ([]*Secret, error)

	// OrgSecretCreate creates an organization secret.
	OrgSecretCreate(owner string, secret *Secret) (*Secret, error)

	// OrgSecretUpdate updates an organization secret.
	OrgSecretUpdate(owner string, secret *Secret) (*Secret, error)

	// OrgSecretDelete deletes an organization secret.
	OrgSecretDelete(owner, secret string) error

	// GlobalSecret returns an global secret by name.
	GlobalSecret(secret string) (*Secret, error)

	// GlobalSecretList returns a list of all global secrets.
	GlobalSecretList() ([]*Secret, error)

	// GlobalSecretCreate creates a global secret.
	GlobalSecretCreate(secret *Secret) (*Secret, error)

	// GlobalSecretUpdate updates a global secret.
	GlobalSecretUpdate(secret *Secret) (*Secret, error)

	// GlobalSecretDelete deletes a global secret.
	GlobalSecretDelete(secret string) error

	// QueueInfo returns the queue state.
	QueueInfo() (*Info, error)

	// LogLevel returns the current logging level
	LogLevel() (*LogLevel, error)

	// SetLogLevel sets the server's logging level
	SetLogLevel(logLevel *LogLevel) (*LogLevel, error)

	// CronList list all cron jobs of a repo
	CronList(owner, repo string) ([]*Cron, error)

	// CronGet get a specific cron job of a repo by id
	CronGet(owner, repo string, cronID int64) (*Cron, error)

	// CronDelete delete a specific cron job of a repo by id
	CronDelete(owner, repo string, cronID int64) error

	// CronCreate create a new cron job in a repo
	CronCreate(owner, repo string, cron *Cron) (*Cron, error)

	// CronUpdate update an existing cron job of a repo
	CronUpdate(owner, repo string, cron *Cron) (*Cron, error)
}
