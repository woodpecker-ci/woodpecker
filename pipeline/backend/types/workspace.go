package types

// BackendOptions defines all the advanced options for the kubernetes backend
type Workspace struct {
	Tmpfs          []Tmpfs `json:"Tmpfs"`
}

type Tmpfs struct {
	Path string `json:"path,omitempty"`
	Size int `json:"size,omitempty"`
}
