package kubectl

import (
	"strings"
)

func (this *KubeCtlBackend) RenderSetupYaml() (string, error) {
	var yamlParts []string

	for _, vol := range this.Config.Volumes {
		volTemplate := KubeVolumeTemplate{
			Engine: this,
			Name:   vol.Name,
		}
		volYaml, err := volTemplate.Render()
		if err != nil {
			return "", err
		}
		yamlParts = append(yamlParts, volYaml)
	}

	return strings.Join(yamlParts, "\n---\n"), nil
}
