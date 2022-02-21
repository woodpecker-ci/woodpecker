package kubectl

import (
	"strings"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type KubeJobTemplate struct {
	Step          *types.Step  // The executing step
	Backend       *KubeBackend // the executing engine
	DetachedPodIP string       // The main pod ip.
}

func (template *KubeJobTemplate) Render() (string, error) {
	return renderTemplate("templates/step_job.yaml", template)
}

func (template *KubeJobTemplate) JobName() string {
	return toKuberenetesValidName(template.Backend.ID()+"-"+template.Step.Name, 60)
}

func (template *KubeJobTemplate) JobID() string {
	return template.Backend.activeRun.RunID + "-" + template.Step.Name
}

func (template *KubeJobTemplate) ShellCommand() string {
	return strings.Join(template.Step.Command, ";")
}

func (template *KubeJobTemplate) HasShellCommand() bool {
	return len(template.Step.Command) != 0
}

func (template *KubeJobTemplate) PullPolicy() string {
	return Triary(template.Step.Pull, "Always", "IfNotPresent").(string)
}

func (template *KubeJobTemplate) DetachedHostAlias() string {
	return Triary(
		len(template.Step.Alias) > 0,
		template.Step.Alias,
		toKuberenetesValidName(template.Step.Name, 50),
	).(string)
}

func (template *KubeJobTemplate) HasDNSCondig() bool {
	return len(template.Step.DNS) > 0 || len(template.Step.DNSSearch) > 0
}

type KubeJobTemplateMount struct {
	MountPath string
	PVC       KubePVCTemplate
}

func (template *KubeJobTemplate) Mounts() []KubeJobTemplateMount {
	mounts := []KubeJobTemplateMount{}
	if template.Step.Detached && !template.Backend.PVCAllowOnDetached {
		// To use detached mounts change the storage class.
		return mounts
	}

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

		if pvc, ok := template.Backend.activeRun.PVCByName[name]; ok {
			mounts = append(mounts, KubeJobTemplateMount{
				MountPath: mountPath,
				PVC:       *pvc,
			})
		}
	}
	return mounts
}
