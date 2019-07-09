package drone

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
		ID          int64  `json:"id,omitempty"`
		Owner       string `json:"owner"`
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		Avatar      string `json:"avatar_url,omitempty"`
		Link        string `json:"link_url,omitempty"`
		Kind        string `json:"scm,omitempty"`
		Clone       string `json:"clone_url,omitempty"`
		Branch      string `json:"default_branch,omitempty"`
		Timeout     int64  `json:"timeout,omitempty"`
		Visibility  string `json:"visibility"`
		IsPrivate   bool   `json:"private,omitempty"`
		IsTrusted   bool   `json:"trusted"`
		IsStarred   bool   `json:"starred,omitempty"`
		IsGated     bool   `json:"gated"`
		AllowPull   bool   `json:"allow_pr"`
		AllowPush   bool   `json:"allow_push"`
		AllowDeploy bool   `json:"allow_deploys"`
		AllowTag    bool   `json:"allow_tags"`
		Config      string `json:"config_file"`
	}

	// RepoPatch defines a repository patch request.
	RepoPatch struct {
		Config       *string `json:"config_file,omitempty"`
		IsTrusted    *bool   `json:"trusted,omitempty"`
		IsGated      *bool   `json:"gated,omitempty"`
		Timeout      *int64  `json:"timeout,omitempty"`
		Visibility   *string `json:"visibility"`
		AllowPull    *bool   `json:"allow_pr,omitempty"`
		AllowPush    *bool   `json:"allow_push,omitempty"`
		AllowDeploy  *bool   `json:"allow_deploy,omitempty"`
		AllowTag     *bool   `json:"allow_tag,omitempty"`
		BuildCounter *int    `json:"build_counter,omitempty"`
	}

	// Build defines a build object.
	Build struct {
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
		Remote    string  `json:"remote"`
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

	// Proc represents a process in the build pipeline.
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
		ID     int64    `json:"id"`
		Name   string   `json:"name"`
		Value  string   `json:"value,omitempty"`
		Images []string `json:"image"`
		Events []string `json:"event"`
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
		Remote   string `json:"remote,omitempty"`
		Title    string `json:"title,omitempty"`
		Message  string `json:"message,omitempty"`
		Author   string `json:"author,omitempty"`
		Avatar   string `json:"author_avatar,omitempty"`
		Email    string `json:"author_email,omitempty"`
	}

	// Server represents a server node.
	Server struct {
		ID       string `json:"id"`
		Provider string `json:"provider"`
		State    string `json:"state"`
		Name     string `json:"name"`
		Image    string `json:"image"`
		Region   string `json:"region"`
		Size     string `json:"size"`
		Address  string `json:"address"`
		Capacity int    `json:"capacity"`
		Secret   string `json:"secret"`
		Error    string `json:"error"`
		CAKey    []byte `json:"ca_key"`
		CACert   []byte `json:"ca_cert"`
		TLSKey   []byte `json:"tls_key"`
		TLSCert  []byte `json:"tls_cert"`
		Created  int64  `json:"created"`
		Updated  int64  `json:"updated"`
		Started  int64  `json:"started"`
		Stopped  int64  `json:"stopped"`
	}

	// Version provides system version details.
	Version struct {
		Source  string `json:"source,omitempty"`
		Version string `json:"version,omitempty"`
		Commit  string `json:"commit,omitempty"`
	}
)
