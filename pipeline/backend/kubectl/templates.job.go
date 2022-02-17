package kubectl

import (
	"strings"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type KubeJobTemplate struct {
	Step    *types.Step     // The executing step
	Backend *KubeCtlBackend // the executing engine
}

func (template *KubeJobTemplate) Render() (string, error) {
	return renderTemplate("templates/step_job.yaml", template)
}

func (template *KubeJobTemplate) JobName() string {
	return toKuberenetesValidName(template.Backend.ID()+"-"+template.Step.Name, 60)
}

func (template *KubeJobTemplate) JobID() string {
	return template.Backend.RunID + "-" + template.Step.Name
}

func (template *KubeJobTemplate) ShellCommand() string {
	return strings.Join(template.Step.Command, ";")
}

type KubeJobTemplateMount struct {
	MountPath string
	PVC       KubePVCTemplate
}

func (template *KubeJobTemplate) Mounts() []KubeJobTemplateMount {
	mounts := []KubeJobTemplateMount{}
	for _, vol := range template.Step.Volumes {
		volArgs := strings.Split(vol, ":")
		if len(volArgs) < 2 {
			continue
		}

		name := strings.TrimSpace(volArgs[0])
		mountPath := strings.TrimSpace(volArgs[1])
		if len(mountPath) == 0 || len(name) == 0 {
			continue
		}

		if pvc, ok := template.Backend.PVCByName[name]; ok {
			mounts = append(mounts, KubeJobTemplateMount{
				MountPath: mountPath,
				PVC:       *pvc,
			})
		}
	}
	return mounts
}
