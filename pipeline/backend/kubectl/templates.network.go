package kubectl

type KubeNetworkPolicyTemplate struct {
	Backend *KubeBackend // the executing engine
}

func (template *KubeNetworkPolicyTemplate) Render() (string, error) {
	return RenderTextTemplate("templates/network_policy.yaml", template)
}
