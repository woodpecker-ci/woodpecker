package kubernetes

import (
	"strings"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/common"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Pod(namespace string, step *types.Step, labels, platform map[string]string, annotations map[string]string) (*v1.Pod, error) {
	var (
		vols       []v1.Volume
		volMounts  []v1.VolumeMount
		entrypoint []string
		args       []string
	)

	if step.WorkingDir != "" {
		for _, vol := range step.Volumes {
			volumeName, err := dnsName(strings.Split(vol, ":")[0])
			if err != nil {
				return nil, err
			}

			vols = append(vols, v1.Volume{
				Name: volumeName,
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: volumeName,
						ReadOnly:  false,
					},
				},
			})

			volMounts = append(volMounts, v1.VolumeMount{
				Name:      volumeName,
				MountPath: volumeMountPath(vol),
			})
		}
	}

	pullPolicy := v1.PullIfNotPresent
	if step.Pull {
		pullPolicy = v1.PullAlways
	}

	if len(step.Commands) != 0 {
		scriptEnv, entry, cmds := common.GenerateContainerConf(step.Commands)
		for k, v := range scriptEnv {
			step.Environment[k] = v
		}
		entrypoint = entry
		args = cmds
	}

	hostAliases := []v1.HostAlias{}
	for _, extraHost := range step.ExtraHosts {
		host := strings.Split(extraHost, ":")
		hostAliases = append(hostAliases, v1.HostAlias{IP: host[1], Hostnames: []string{host[0]}})
	}

	// TODO: add support for resource limits
	// if step.Resources.CPULimit == "" {
	// 	step.Resources.CPULimit = "2"
	// }
	// if step.Resources.MemoryLimit == "" {
	// 	step.Resources.MemoryLimit = "2G"
	// }
	// memoryLimit := resource.MustParse(step.Resources.MemoryLimit)
	// CPULimit := resource.MustParse(step.Resources.CPULimit)

	memoryLimit := resource.MustParse("2G")
	CPULimit := resource.MustParse("2")

	memoryLimitValue, _ := memoryLimit.AsInt64()
	CPULimitValue, _ := CPULimit.AsInt64()
	loadfactor := 0.5

	memoryRequest := resource.NewQuantity(int64(float64(memoryLimitValue)*loadfactor), resource.DecimalSI)
	CPURequest := resource.NewQuantity(int64(float64(CPULimitValue)*loadfactor), resource.DecimalSI)

	resources := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: *memoryRequest,
			v1.ResourceCPU:    *CPURequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory: memoryLimit,
			v1.ResourceCPU:    CPULimit,
		},
	}

	podName, err := dnsName(step.Name)
	if err != nil {
		return nil, err
	}

	labels["step"] = podName

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        podName,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			HostAliases:   hostAliases,
			NodeSelector:  platform,
			Containers: []v1.Container{{
				Name:            podName,
				Image:           step.Image,
				ImagePullPolicy: pullPolicy,
				Command:         entrypoint,
				Args:            args,
				WorkingDir:      step.WorkingDir,
				Env:             mapToEnvVars(step.Environment),
				VolumeMounts:    volMounts,
				Resources:       resources,
				SecurityContext: &v1.SecurityContext{
					Privileged: &step.Privileged,
				},
			}},
			ImagePullSecrets: []v1.LocalObjectReference{{Name: "regcred"}},
			Volumes:          vols,
		},
	}

	return pod, nil
}

func mapToEnvVars(m map[string]string) []v1.EnvVar {
	var ev []v1.EnvVar
	for k, v := range m {
		ev = append(ev, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return ev
}

func volumeMountPath(i string) string {
	s := strings.Split(i, ":")
	if len(s) > 1 {
		return s[1]
	}
	return s[0]
}
