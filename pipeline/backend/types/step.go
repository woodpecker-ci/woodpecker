package types

// Step defines a container process.
type Step struct {
	Name         string            `json:"name"`
	Alias        string            `json:"alias,omitempty"`
	Image        string            `json:"image,omitempty"`
	Pull         bool              `json:"pull,omitempty"`
	Detached     bool              `json:"detach,omitempty"`
	Privileged   bool              `json:"privileged,omitempty"`
	WorkingDir   string            `json:"working_dir,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Entrypoint   []string          `json:"entrypoint,omitempty"`
	Commands     []string          `json:"commands,omitempty"`
	ExtraHosts   []string          `json:"extra_hosts,omitempty"`
	Volumes      []string          `json:"volumes,omitempty"`
	Tmpfs        []string          `json:"tmpfs,omitempty"`
	Devices      []string          `json:"devices,omitempty"`
	Networks     []Conn            `json:"networks,omitempty"`
	DNS          []string          `json:"dns,omitempty"`
	DNSSearch    []string          `json:"dns_search,omitempty"`
	MemSwapLimit int64             `json:"memswap_limit,omitempty"`
	MemLimit     int64             `json:"mem_limit,omitempty"`
	ShmSize      int64             `json:"shm_size,omitempty"`
	CPUQuota     int64             `json:"cpu_quota,omitempty"`
	CPUShares    int64             `json:"cpu_shares,omitempty"`
	CPUSet       string            `json:"cpu_set,omitempty"`
	OnFailure    bool              `json:"on_failure,omitempty"`
	OnSuccess    bool              `json:"on_success,omitempty"`
	Failure      string            `json:"failure,omitempty"`
	AuthConfig   Auth              `json:"auth_config,omitempty"`
	NetworkMode  string            `json:"network_mode,omitempty"`
	IpcMode      string            `json:"ipc_mode,omitempty"`
	Sysctls      map[string]string `json:"sysctls,omitempty"`
}
