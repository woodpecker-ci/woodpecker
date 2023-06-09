package types

// KubernetesBackendOptions defines all the advanced options for the kubernetes backend
type KubernetesBackendOptions struct {
	Resources Resources `json:"resouces,omitempty"`
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// Resources defines two maps for kubernetes resource definitions
type Resources struct {
	Requests map[string]string `json:"requests,omitempty"`
	Limits   map[string]string `json:"limits,omitempty"`
}
