package kubectl

type KubePVCTemplate struct {
	StorageClassName string
	StorageSize      string
	Name             string
	Backend          *KubeBackend // the executing engine
}

func (template *KubePVCTemplate) VolumeName() string {
	return toKuberenetesValidName(template.Backend.ID()+"-"+template.Name, 60)
}

func (template *KubePVCTemplate) MountName() string {
	return toKuberenetesValidName(template.Name, 60)
}

func (template *KubePVCTemplate) Render() (string, error) {
	return renderTemplate("templates/volume_claim.yaml", template)
}
