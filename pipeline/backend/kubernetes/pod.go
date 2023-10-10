// Copyright 2022 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubernetes

import (
	"fmt"
	"maps"
	"strings"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/common"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
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

	var pullPolicy v1.PullPolicy
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

	var serviceAccountName string
	if step.BackendOptions.Kubernetes.ServiceAccountName != "" {
		serviceAccountName = step.BackendOptions.Kubernetes.ServiceAccountName
	}

	podName, err := dnsName(step.Name)
	if err != nil {
		return nil, err
	}

	labels["step"] = podName

	var nodeSelector map[string]string
	platform, exist := step.Environment["CI_SYSTEM_PLATFORM"]
	if exist && platform != "" {
		arch := strings.Split(platform, "/")[1]
		nodeSelector = map[string]string{v1.LabelArchStable: arch}
		log.Trace().Msgf("Using the node selector from the Agent's platform: %v", nodeSelector)
	}
	beOptNodeSelector := step.BackendOptions.Kubernetes.NodeSelector
	if len(beOptNodeSelector) > 0 {
		if len(nodeSelector) == 0 {
			nodeSelector = beOptNodeSelector
		} else {
			log.Trace().Msgf("Appending labels to the node selector from the backend options: %v", beOptNodeSelector)
			maps.Copy(nodeSelector, beOptNodeSelector)
		}
	}

	var tolerations []v1.Toleration
	beTolerations := step.BackendOptions.Kubernetes.Tolerations
	if len(beTolerations) > 0 {
		for _, t := range step.BackendOptions.Kubernetes.Tolerations {
			toleration := v1.Toleration{
				Key:               t.Key,
				Operator:          v1.TolerationOperator(t.Operator),
				Value:             t.Value,
				Effect:            v1.TaintEffect(t.Effect),
				TolerationSeconds: t.TolerationSeconds,
			}
			tolerations = append(tolerations, toleration)
		}
		log.Trace().Msgf("Tolerations that will be used in the backend options: %v", beTolerations)
	}

	beSecurityContext := step.BackendOptions.Kubernetes.SecurityContext
	log.Trace().Interface("Security context", beSecurityContext).Msg("Security context that will be used for pods/containers")
	podSecCtx := podSecurityContext(beSecurityContext)
	containerSecCtx := containerSecurityContext(beSecurityContext, step.Privileged)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        podName,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1.PodSpec{
			RestartPolicy:      v1.RestartPolicyNever,
			HostAliases:        hostAliases,
			NodeSelector:       nodeSelector,
			Tolerations:        tolerations,
			ServiceAccountName: serviceAccountName,
			SecurityContext:    podSecCtx,
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
				SecurityContext: containerSecCtx,
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

func podSecurityContext(sc *types.SecurityContext) *v1.PodSecurityContext {
	if sc != nil {
		return &v1.PodSecurityContext{
			RunAsNonRoot: sc.RunAsNonRoot,
			RunAsUser:    sc.RunAsUser,
			RunAsGroup:   sc.RunAsGroup,
			FSGroup:      sc.FSGroup,
		}
	}
	return nil
}

func containerSecurityContext(sc *types.SecurityContext, privileged bool) *v1.SecurityContext {
	containerSecCtx := &v1.SecurityContext{
		Privileged: &privileged,
	}

	if sc != nil {
		if sc.Privileged != nil {
			privileged = privileged || *sc.Privileged
			containerSecCtx.Privileged = &privileged
		}
	}

	return containerSecCtx
}
