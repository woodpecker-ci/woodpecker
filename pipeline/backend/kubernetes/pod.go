package kubernetes

import (
	"fmt"
	"strings"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/common"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Pod(namespace string, step *types.Step, labels, annotations map[string]string) (*v1.Pod, error) {
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

	resourceRequirements := v1.ResourceRequirements{Requests: v1.ResourceList{}, Limits: v1.ResourceList{}}
	var err error
	for key, val := range step.BackendOptions.Kubernetes.Resources.Requests {
		resourceKey := v1.ResourceName(key)
		resourceRequirements.Requests[resourceKey], err = resource.ParseQuantity(val)
		if err != nil {
			return nil, fmt.Errorf("resource request '%v' quantity '%v': %w", key, val, err)
		}
	}
	for key, val := range step.BackendOptions.Kubernetes.Resources.Limits {
		resourceKey := v1.ResourceName(key)
		resourceRequirements.Limits[resourceKey], err = resource.ParseQuantity(val)
		if err != nil {
			return nil, fmt.Errorf("resource limit '%v' quantity '%v': %w", key, val, err)
		}
	}

	podName, err := dnsName(step.Name)
	if err != nil {
		return nil, err
	}

	labels["step"] = podName

	var platform string
	for _, e := range mapToEnvVars(step.Environment) {
		if e.Name == "CI_SYSTEM_ARCH" {
			platform = e.Value
			break
		}
	}

	NodeSelector := map[string]string{"kubernetes.io/arch": strings.Split(platform, "/")[1]}

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
			NodeSelector:  NodeSelector,
			Containers: []v1.Container{{
				Name:            podName,
				Image:           step.Image,
				ImagePullPolicy: pullPolicy,
				Command:         entrypoint,
				Args:            args,
				WorkingDir:      step.WorkingDir,
				Env:             mapToEnvVars(step.Environment),
				VolumeMounts:    volMounts,
				Resources:       resourceRequirements,
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
