package types

// BackendOptions defines advanced options for specific backends
type BackendOptions struct {
	Kubernetes KubernetesBackendOptions `json:"kubernetes,omitempty"`
}
