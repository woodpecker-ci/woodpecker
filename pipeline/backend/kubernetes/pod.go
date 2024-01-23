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
		Name:      podName,
		Namespace: config.Namespace,
	}

	meta.Labels = config.PodLabels
	if meta.Labels == nil {
		meta.Labels = make(map[string]string, 1)
	}
	meta.Labels[StepLabel] = step.Name

	if step.Type == types.StepTypeService {
		meta.Labels[ServiceLabel] = step.Name
	}

	meta.Annotations = config.PodAnnotations
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}

	securityContext := step.BackendOptions.Kubernetes.SecurityContext
	if securityContext != nil {
		key, value := apparmorAnnotation(podName, securityContext.ApparmorProfile)
		if key != nil && value != nil {
			meta.Annotations[*key] = *value
		}
	}

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
		if len(step.Entrypoint) > 0 {
			command = step.Entrypoint
		}
		container.Command = command
		container.Args = []string{args}
		maps.Copy(step.Environment, scriptEnv)
	}

	container.Env = mapToEnvVars(step.Environment)
	container.Ports = containerPorts(step.Ports)
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

func containerPorts(ports []types.Port) []v1.ContainerPort {
	containerPorts := make([]v1.ContainerPort, len(ports))
	for i, port := range ports {
		containerPorts[i] = containerPort(port)
	}
	return containerPorts
}

func containerPort(port types.Port) v1.ContainerPort {
	return v1.ContainerPort{
		ContainerPort: int32(port.Number),
		Protocol:      v1.Protocol(strings.ToUpper(port.Protocol)),
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
	log.Trace().Msgf("using the image pull secrets: %v", imagePullSecretNames)

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
			return nil, fmt.Errorf("resource request '%s' quantity '%s': %w", key, val, err)
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
		log.Trace().Msgf("using the node selector from the Agent's platform: %v", nodeSelector)
	}

	if len(backendNodeSelector) > 0 {
		log.Trace().Msgf("appending labels to the node selector from the backend options: %v", backendNodeSelector)
		maps.Copy(nodeSelector, backendNodeSelector)
	}

	return nodeSelector
}

func tolerations(backendTolerations []types.Toleration) []v1.Toleration {
	var tolerations []v1.Toleration

	if len(backendTolerations) > 0 {
		log.Trace().Msgf("tolerations that will be used in the backend options: %v", backendTolerations)
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
		seccomp *v1.SeccompProfile
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

	if sc != nil {
		seccomp = seccompProfile(sc.SeccompProfile)
	}

	if nonRoot == nil && user == nil && group == nil && fsGroup == nil && seccomp == nil {
		return nil
	}

	securityContext := &v1.PodSecurityContext{
		RunAsNonRoot:   nonRoot,
		RunAsUser:      user,
		RunAsGroup:     group,
		FSGroup:        fsGroup,
		SeccompProfile: seccomp,
	}
	log.Trace().Msgf("pod security context that will be used: %v", securityContext)
	return securityContext
}

func seccompProfile(scp *types.SecProfile) *v1.SeccompProfile {
	if scp == nil || len(scp.Type) == 0 {
		return nil
	}
	log.Trace().Msgf("using seccomp profile: %v", scp)

	seccompProfile := &v1.SeccompProfile{
		Type: v1.SeccompProfileType(scp.Type),
	}
	if len(scp.LocalhostProfile) > 0 {
		seccompProfile.LocalhostProfile = &scp.LocalhostProfile
	}

	return seccompProfile
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
	log.Trace().Msgf("container security context that will be used: %v", securityContext)
	return securityContext
}

func apparmorAnnotation(containerName string, scp *types.SecProfile) (*string, *string) {
	if scp == nil {
		return nil, nil
	}
	log.Trace().Msgf("using AppArmor profile: %v", scp)

	var (
		profileType string
		profilePath string
	)

	if scp.Type == types.SecProfileTypeRuntimeDefault {
		profileType = "runtime"
		profilePath = "default"
	}

	if scp.Type == types.SecProfileTypeLocalhost {
		profileType = "localhost"
		profilePath = scp.LocalhostProfile
	}

	if len(profileType) == 0 {
		return nil, nil
	}

	key := v1.AppArmorBetaContainerAnnotationKeyPrefix + containerName
	value := profileType + "/" + profilePath
	return &key, &value
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
	podName, err := stepToPodName(step)
	if err != nil {
		return nil, err
	}
	engineConfig := engine.getConfig()
	pod, err := mkPod(step, engineConfig, podName, engine.goos)
	if err != nil {
		return nil, err
	}

	log.Trace().Msgf("creating pod: %s", pod.Name)
	return engine.client.CoreV1().Pods(engineConfig.Namespace).Create(ctx, pod, metav1.CreateOptions{})
}

func stopPod(ctx context.Context, engine *kube, step *types.Step, deleteOpts metav1.DeleteOptions) error {
	podName, err := stepToPodName(step)
	if err != nil {
		return err
	}
	log.Trace().Str("name", podName).Msg("deleting pod")

	err = engine.client.CoreV1().Pods(engine.config.Namespace).Delete(ctx, podName, deleteOpts)
	if errors.IsNotFound(err) {
		// Don't abort on 404 errors from k8s, they most likely mean that the pod hasn't been created yet, usually because pipeline was canceled before running all steps.
		return nil
	}
	return err
}
