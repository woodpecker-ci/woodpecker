package kubectl

type KubePVCTemplate struct {
	Name string
	Run  *KubeBackendRun // the executing engine
}

// The pvc volume name.
func (template *KubePVCTemplate) VolumeName() string {
	return ToKuberenetesValidName(template.Run.ID()+"-"+template.Name, 60)
}

// The pvc mount name.
func (template *KubePVCTemplate) MountName() string {
	return ToKuberenetesValidName(template.Name, 60)
}

func (template *KubePVCTemplate) Render() (string, error) {
	return RenderTextTemplate("templates/volume_claim.yaml", template)
}
