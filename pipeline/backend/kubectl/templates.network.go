package kubectl

type KubeNetworkPolicyTemplate struct {
	Backend   *KubeCtlBackend // the executing engine
	DenyCIDR  []string        // Deny access to CIDR
	AllowCIDR []string        // Allow access to cidr
}

func (template *KubeNetworkPolicyTemplate) Render() (string, error) {
	return renderTemplate("templates/network_policy.yaml", template)
}
