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
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/common"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

const (
	StepLabel = "step"
	podPrefix = "wp-"
)

func mkPod(step *types.Step, config *config, podName, goos string) (*v1.Pod, error) {
	meta := podMeta(step, config, podName)

	spec, err := podSpec(step, config)
	if err != nil {
		return nil, err
	}

	container, err := podContainer(step, podName, goos)
	if err != nil {
		return nil, err
	}
	spec.Containers = append(spec.Containers, container)

	pod := &v1.Pod{
		ObjectMeta: meta,
		Spec:       spec,
	}

	return pod, nil
}

func stepToPodName(step *types.Step) (name string, err error) {
	if step.Type == types.StepTypeService {
		return serviceName(step)
	}
	return podName(step)
}

func podName(step *types.Step) (string, error) {
	return dnsName(podPrefix + step.UUID)
}

func podMeta(step *types.Step, config *config, podName string) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:        podName,
		Namespace:   config.Namespace,
		Annotations: config.PodAnnotations,
	}

	labels := make(map[string]string, len(config.PodLabels)+1)
	// copy to not alter the engine config
	maps.Copy(labels, config.PodLabels)
	labels[StepLabel] = step.Name
	meta.Labels = labels

	return meta
}

func podSpec(step *types.Step, config *config) (v1.PodSpec, error) {
	var err error
	spec := v1.PodSpec{
		RestartPolicy:      v1.RestartPolicyNever,
		ServiceAccountName: step.BackendOptions.Kubernetes.ServiceAccountName,
		ImagePullSecrets:   imagePullSecretsReferences(config.ImagePullSecretNames),
		HostAliases:        hostAliases(step.ExtraHosts),
		NodeSelector:       nodeSelector(step.BackendOptions.Kubernetes.NodeSelector, step.Environment["CI_SYSTEM_PLATFORM"]),
		Tolerations:        tolerations(step.BackendOptions.Kubernetes.Tolerations),
		SecurityContext:    podSecurityContext(step.BackendOptions.Kubernetes.SecurityContext, config.SecurityContext),
	}
	spec.Volumes, err = volumes(step.Volumes)
	if err != nil {
		return spec, err
	}

	return spec, nil
}

func podContainer(step *types.Step, podName, goos string) (v1.Container, error) {
	var err error
	container := v1.Container{
		Name:       podName,
		Image:      step.Image,
		WorkingDir: step.WorkingDir,
	}

	if step.Pull {
		container.ImagePullPolicy = v1.PullAlways
	}

	if len(step.Commands) != 0 {
		scriptEnv, command, args := common.GenerateContainerConf(step.Commands, goos)
		container.Command = command
		container.Args = args
		maps.Copy(step.Environment, scriptEnv)
	}

	container.Env = mapToEnvVars(step.Environment)
	container.SecurityContext = containerSecurityContext(step.BackendOptions.Kubernetes.SecurityContext, step.Privileged)

	container.Resources, err = resourceRequirements(step.BackendOptions.Kubernetes.Resources)
	if err != nil {
		return container, err
	}

	container.VolumeMounts, err = volumeMounts(step.Volumes)
	if err != nil {
		return container, err
	}

	return container, nil
}

func volumes(volumes []string) ([]v1.Volume, error) {
	var vols []v1.Volume

	for _, v := range volumes {
		volumeName, err := volumeName(v)
		if err != nil {
			return nil, err
		}
		vols = append(vols, volume(volumeName))
	}

	return vols, nil
}

func volume(name string) v1.Volume {
	pvcSource := v1.PersistentVolumeClaimVolumeSource{
		ClaimName: name,
		ReadOnly:  false,
	}
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &pvcSource,
		},
	}
}

func volumeMounts(volumes []string) ([]v1.VolumeMount, error) {
	var mounts []v1.VolumeMount

	for _, v := range volumes {
		volumeName, err := volumeName(v)
		if err != nil {
			return nil, err
		}

		mount := volumeMount(volumeName, volumeMountPath(v))
		mounts = append(mounts, mount)
	}
	return mounts, nil
}

func volumeMount(name, path string) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      name,
		MountPath: path,
	}
}

// Here is the service IPs (placed in /etc/hosts in the Pod)
func hostAliases(extraHosts []types.HostAlias) []v1.HostAlias {
	hostAliases := []v1.HostAlias{}
	for _, extraHost := range extraHosts {
		hostAlias := hostAlias(extraHost)
		hostAliases = append(hostAliases, hostAlias)
	}
	return hostAliases
}

func hostAlias(extraHost types.HostAlias) v1.HostAlias {
	return v1.HostAlias{
		IP:        extraHost.IP,
		Hostnames: []string{extraHost.Name},
	}
}

func imagePullSecretsReferences(imagePullSecretNames []string) []v1.LocalObjectReference {
	log.Trace().Msgf("Using the image pull secrets: %v", imagePullSecretNames)

	secretReferences := make([]v1.LocalObjectReference, len(imagePullSecretNames))
	for i, imagePullSecretName := range imagePullSecretNames {
		secretReferences[i] = imagePullSecretsReference(imagePullSecretName)
	}
	return secretReferences
}

func imagePullSecretsReference(imagePullSecretName string) v1.LocalObjectReference {
	return v1.LocalObjectReference{
		Name: imagePullSecretName,
	}
}

func resourceRequirements(resources types.Resources) (v1.ResourceRequirements, error) {
	var err error
	requirements := v1.ResourceRequirements{}

	requirements.Requests, err = resourceList(resources.Requests)
	if err != nil {
		return requirements, err
	}

	requirements.Limits, err = resourceList(resources.Limits)
	if err != nil {
		return requirements, err
	}

	return requirements, nil
}

func resourceList(resources map[string]string) (v1.ResourceList, error) {
	requestResources := v1.ResourceList{}
	for key, val := range resources {
		resName := v1.ResourceName(key)
		resVal, err := resource.ParseQuantity(val)
		if err != nil {
			return nil, fmt.Errorf("resource request '%v' quantity '%v': %w", key, val, err)
		}
		requestResources[resName] = resVal
	}
	return requestResources, nil
}

func nodeSelector(backendNodeSelector map[string]string, platform string) map[string]string {
	nodeSelector := make(map[string]string)

	if platform != "" {
		arch := strings.Split(platform, "/")[1]
		nodeSelector[v1.LabelArchStable] = arch
		log.Trace().Msgf("Using the node selector from the Agent's platform: %v", nodeSelector)
	}

	if len(backendNodeSelector) > 0 {
		log.Trace().Msgf("Appending labels to the node selector from the backend options: %v", backendNodeSelector)
		maps.Copy(nodeSelector, backendNodeSelector)
	}

	return nodeSelector
}

func tolerations(backendTolerations []types.Toleration) []v1.Toleration {
	var tolerations []v1.Toleration

	if len(backendTolerations) > 0 {
		log.Trace().Msgf("Tolerations that will be used in the backend options: %v", backendTolerations)
		for _, backendToleration := range backendTolerations {
			toleration := toleration(backendToleration)
			tolerations = append(tolerations, toleration)
		}
	}

	return tolerations
}

func toleration(backendToleration types.Toleration) v1.Toleration {
	return v1.Toleration{
		Key:               backendToleration.Key,
		Operator:          v1.TolerationOperator(backendToleration.Operator),
		Value:             backendToleration.Value,
		Effect:            v1.TaintEffect(backendToleration.Effect),
		TolerationSeconds: backendToleration.TolerationSeconds,
	}
}

func podSecurityContext(sc *types.SecurityContext, secCtxConf SecurityContextConfig) *v1.PodSecurityContext {
	var (
		nonRoot *bool
		user    *int64
		group   *int64
		fsGroup *int64
	)

	if sc != nil && sc.RunAsNonRoot != nil {
		if *sc.RunAsNonRoot {
			nonRoot = sc.RunAsNonRoot // true
		}
	} else if secCtxConf.RunAsNonRoot {
		nonRoot = &secCtxConf.RunAsNonRoot // true
	}

	if sc != nil {
		user = sc.RunAsUser
		group = sc.RunAsGroup
		fsGroup = sc.FSGroup
	}

	if nonRoot == nil && user == nil && group == nil && fsGroup == nil {
		return nil
	}

	securityContext := &v1.PodSecurityContext{
		RunAsNonRoot: nonRoot,
		RunAsUser:    user,
		RunAsGroup:   group,
		FSGroup:      fsGroup,
	}
	log.Trace().Msgf("Pod security context that will be used: %v", securityContext)
	return securityContext
}

func containerSecurityContext(sc *types.SecurityContext, stepPrivileged bool) *v1.SecurityContext {
	var privileged *bool

	if sc != nil && sc.Privileged != nil && *sc.Privileged {
		privileged = sc.Privileged // true
	} else if stepPrivileged {
		privileged = &stepPrivileged // true
	}

	if privileged == nil {
		return nil
	}

	securityContext := &v1.SecurityContext{
		Privileged: privileged,
	}
	log.Trace().Msgf("Container security context that will be used: %v", securityContext)
	return securityContext
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

func startPod(ctx context.Context, engine *kube, step *types.Step) (*v1.Pod, error) {
	podName, err := podName(step)
	if err != nil {
		return nil, err
	}
	pod, err := mkPod(step, engine.config, podName, engine.goos)
	if err != nil {
		return nil, err
	}

	log.Trace().Msgf("Creating pod: %s", pod.Name)
	return engine.client.CoreV1().Pods(engine.config.Namespace).Create(ctx, pod, metav1.CreateOptions{})
}

func stopPod(ctx context.Context, engine *kube, step *types.Step, deleteOpts metav1.DeleteOptions) error {
	podName, err := podName(step)
	if err != nil {
		return err
	}
	log.Trace().Str("name", podName).Msg("Deleting pod")

	err = engine.client.CoreV1().Pods(engine.config.Namespace).Delete(ctx, podName, deleteOpts)
	if errors.IsNotFound(err) {
		// Don't abort on 404 errors from k8s, they most likely mean that the pod hasn't been created yet, usually because pipeline was canceled before running all steps.
		return nil
	}
	return err
}
