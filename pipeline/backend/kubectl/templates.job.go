package kubectl

import (
	"strings"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type KubeJobTemplate struct {
	Step          *types.Step     // The executing step
	Run           *KubePiplineRun // the executing engine
	DetachedPodIP string          // The main pod ip.
}

func (template *KubeJobTemplate) Render() (string, error) {
	return RenderTextTemplate("templates/step_job.yaml", template)
}

// The job kubernetes name.
func (template *KubeJobTemplate) JobName() string {
	// This name should be unique (Run ID is random and unique)
	return ToKubernetesValidName(template.Run.ID()+"-"+template.Step.Name, 60)
}

func (template *KubeJobTemplate) KubeLabels() map[string]string {
	kubeLabels := map[string]string{}
	for k, v := range template.Step.Labels {
		kubeLabels[ToKubernetesValidName(k, 30)] = ToKubernetesValidName(v, -1)
	}
	return kubeLabels
}

// The job id
func (template *KubeJobTemplate) JobID() string {
	return ToKubernetesValidName(template.Run.RunID+"-"+template.Step.Name, -1)
}

// If true a shell command exists.
func (template *KubeJobTemplate) HasShellCommand() bool {
	return len(template.Step.Command) != 0
}

// The shell command to execute the job (from the step)
func (template *KubeJobTemplate) ShellCommand() string {
	return strings.Join(template.Step.Command, ";")
}

// The active kubernetes pull policy.
func (template *KubeJobTemplate) PullPolicy() string {
	if len(template.Run.Backend.ForcePullPolicy) > 0 {
		return template.Run.Backend.ForcePullPolicy
	}
	return Triary(template.Step.Pull, "Always", "IfNotPresent").(string)
}

// The alias name for the current job.
func (template *KubeJobTemplate) DetachedHostAlias() string {
	return Triary(
		len(template.Step.Alias) > 0,
		template.Step.Alias,
		ToKubernetesValidName(template.Step.Name, 50),
	).(string)
}

// If true, has a DNS config.
func (template *KubeJobTemplate) HasDNSCondig() bool {
	return len(template.Step.DNS) > 0 || len(template.Step.DNSSearch) > 0
}

type KubeJobTemplateMount struct {
	MountPath string
	PVC       KubePVCTemplate
}

// A list of mounts for the current job.
func (template *KubeJobTemplate) Mounts() []KubeJobTemplateMount {
	mounts := []KubeJobTemplateMount{}
	if template.Step.Detached && !template.Run.Backend.PVCAllowOnDetached {
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

		if pvc, ok := template.Run.PVCByName[name]; ok {
			mounts = append(mounts, KubeJobTemplateMount{
				MountPath: mountPath,
				PVC:       *pvc,
			})
		}
	}
	return mounts
}
