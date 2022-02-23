package kubectl

type KubeNetworkPolicyTemplate struct {
	Run *KubePiplineRun // the executing engine
}

func (template *KubeNetworkPolicyTemplate) Render() (string, error) {
	return RenderTextTemplate("templates/network_policy.yaml", template)
}
