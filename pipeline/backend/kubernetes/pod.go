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
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/common"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

const (
	StepLabel = "step"
	podPrefix = "wp-"
)

func mkPod(step *types.Step, config *config, podName, goos string, options BackendOptions) (*v1.Pod, error) {
	var err error

	nsp := newNativeSecretsProcessor(config, options.Secrets)
	err = nsp.process()
	if err != nil {
		return nil, err
	}

	meta, err := podMeta(step, config, options, podName)
	if err != nil {
		return nil, err
	}

	spec, err := podSpec(step, config, options, nsp)
	if err != nil {
		return nil, err
	}

	container, err := podContainer(step, podName, goos, options, nsp)
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

func podMeta(step *types.Step, config *config, options BackendOptions, podName string) (meta_v1.ObjectMeta, error) {
	var err error
	meta := meta_v1.ObjectMeta{
		Name:        podName,
		Namespace:   config.Namespace,
		Annotations: podAnnotations(config, options),
	}

	meta.Labels, err = podLabels(step, config, options)
	if err != nil {
		return meta, err
	}

	return meta, nil
}

func podLabels(step *types.Step, config *config, options BackendOptions) (map[string]string, error) {
	var err error
	labels := make(map[string]string)

	if len(options.Labels) > 0 {
		if config.PodLabelsAllowFromStep {
			log.Trace().Msgf("using labels from the backend options: %v", options.Labels)
			maps.Copy(labels, options.Labels)
		} else {
			log.Debug().Msg("Pod labels were defined in backend options, but its using disallowed by instance configuration")
		}
	}
	if len(config.PodLabels) > 0 {
		log.Trace().Msgf("using labels from the configuration: %v", config.PodLabels)
		maps.Copy(labels, config.PodLabels)
	}
	if step.Type == types.StepTypeService {
		labels[ServiceLabel], _ = serviceName(step)
	}
	labels[StepLabel], err = stepLabel(step)
	if err != nil {
		return labels, err
	}

	return labels, nil
}

func stepLabel(step *types.Step) (string, error) {
	return toDNSName(step.Name)
}

func podAnnotations(config *config, options BackendOptions) map[string]string {
	annotations := make(map[string]string)

	if len(options.Annotations) > 0 {
		if config.PodAnnotationsAllowFromStep {
			log.Trace().Msgf("using annotations from the backend options: %v", options.Annotations)
			maps.Copy(annotations, options.Annotations)
		} else {
			log.Debug().Msg("Pod annotations were defined in backend options, but its using disallowed by instance configuration ")
		}
	}
	if len(config.PodAnnotations) > 0 {
		log.Trace().Msgf("using annotations from the configuration: %v", config.PodAnnotations)
		maps.Copy(annotations, config.PodAnnotations)
	}

	return annotations
}

func podSpec(step *types.Step, config *config, options BackendOptions, nsp nativeSecretsProcessor) (v1.PodSpec, error) {
	var err error
	spec := v1.PodSpec{
		RestartPolicy:      v1.RestartPolicyNever,
		RuntimeClassName:   options.RuntimeClassName,
		ServiceAccountName: options.ServiceAccountName,
		HostAliases:        hostAliases(step.ExtraHosts),
		NodeSelector:       nodeSelector(options.NodeSelector, config.PodNodeSelector, step.Environment["CI_SYSTEM_PLATFORM"]),
		Tolerations:        tolerations(options.Tolerations),
		SecurityContext:    podSecurityContext(options.SecurityContext, config.SecurityContext, step.Privileged),
	}
	spec.Volumes, err = pvcVolumes(step.Volumes)
	if err != nil {
		return spec, err
	}

	log.Trace().Msgf("using the image pull secrets: %v", config.ImagePullSecretNames)
	spec.ImagePullSecrets = secretsReferences(config.ImagePullSecretNames)
	if needsRegistrySecret(step) {
		log.Trace().Msgf("using an image pull secret from registries")
		name, err := registrySecretName(step)
		if err != nil {
			return spec, err
		}
		spec.ImagePullSecrets = append(spec.ImagePullSecrets, secretReference(name))
	}

	spec.Volumes = append(spec.Volumes, nsp.volumes...)

	return spec, nil
}

func podContainer(step *types.Step, podName, goos string, options BackendOptions, nsp nativeSecretsProcessor) (v1.Container, error) {
	var err error
	container := v1.Container{
		Name:            podName,
		Image:           step.Image,
		WorkingDir:      step.WorkingDir,
		Ports:           containerPorts(step.Ports),
		SecurityContext: containerSecurityContext(options.SecurityContext, step.Privileged),
	}

	if step.Pull {
		container.ImagePullPolicy = v1.PullAlways
	}

	if len(step.Commands) > 0 {
		scriptEnv, command := common.GenerateContainerConf(step.Commands, goos)
		container.Command = command
		maps.Copy(step.Environment, scriptEnv)
	}
	if len(step.Entrypoint) > 0 {
		container.Command = step.Entrypoint
	}

	container.Env = mapToEnvVars(step.Environment)

	container.Resources, err = resourceRequirements(options.Resources)
	if err != nil {
		return container, err
	}

	container.VolumeMounts, err = volumeMounts(step.Volumes)
	if err != nil {
		return container, err
	}

	container.EnvFrom = append(container.EnvFrom, nsp.envFromSources...)
	container.Env = append(container.Env, nsp.envVars...)
	container.VolumeMounts = append(container.VolumeMounts, nsp.mounts...)

	return container, nil
}

func pvcVolumes(volumes []string) ([]v1.Volume, error) {
	var vols []v1.Volume

	for _, v := range volumes {
		volumeName, err := volumeName(v)
		if err != nil {
			return nil, err
		}
		vols = append(vols, pvcVolume(volumeName))
	}

	return vols, nil
}

func pvcVolume(name string) v1.Volume {
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

// Here is the service IPs (placed in /etc/hosts in the Pod).
func hostAliases(extraHosts []types.HostAlias) []v1.HostAlias {
	var hostAliases []v1.HostAlias
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

func resourceRequirements(resources Resources) (v1.ResourceRequirements, error) {
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

func nodeSelector(backendNodeSelector, configNodeSelector map[string]string, platform string) map[string]string {
	nodeSelector := make(map[string]string)

	if platform != "" {
		arch := strings.Split(platform, "/")[1]
		nodeSelector[v1.LabelArchStable] = arch
		log.Trace().Msgf("using the node selector from the Agent's platform: %v", nodeSelector)
	}

	if len(configNodeSelector) > 0 {
		log.Trace().Msgf("appending labels to the node selector from the configuration: %v", configNodeSelector)
		maps.Copy(nodeSelector, configNodeSelector)
	}

	if len(backendNodeSelector) > 0 {
		log.Trace().Msgf("appending labels to the node selector from the backend options: %v", backendNodeSelector)
		maps.Copy(nodeSelector, backendNodeSelector)
	}

	return nodeSelector
}

func tolerations(backendTolerations []Toleration) []v1.Toleration {
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

func toleration(backendToleration Toleration) v1.Toleration {
	return v1.Toleration{
		Key:               backendToleration.Key,
		Operator:          v1.TolerationOperator(backendToleration.Operator),
		Value:             backendToleration.Value,
		Effect:            v1.TaintEffect(backendToleration.Effect),
		TolerationSeconds: backendToleration.TolerationSeconds,
	}
}

func podSecurityContext(sc *SecurityContext, secCtxConf SecurityContextConfig, stepPrivileged bool) *v1.PodSecurityContext {
	var (
		nonRoot  *bool
		user     *int64
		group    *int64
		fsGroup  *int64
		seccomp  *v1.SeccompProfile
		apparmor *v1.AppArmorProfile
	)

	if secCtxConf.RunAsNonRoot {
		nonRoot = newBool(true)
	}

	if sc != nil {
		// only allow to set user if its not root or step is privileged
		if sc.RunAsUser != nil && (*sc.RunAsUser != 0 || stepPrivileged) {
			user = sc.RunAsUser
		}

		// only allow to set group if its not root or step is privileged
		if sc.RunAsGroup != nil && (*sc.RunAsGroup != 0 || stepPrivileged) {
			group = sc.RunAsGroup
		}

		// only allow to set fsGroup if its not root or step is privileged
		if sc.FSGroup != nil && (*sc.FSGroup != 0 || stepPrivileged) {
			fsGroup = sc.FSGroup
		}

		// only allow to set nonRoot if it's not set globally already
		if nonRoot == nil && sc.RunAsNonRoot != nil {
			nonRoot = sc.RunAsNonRoot
		}

		seccomp = seccompProfile(sc.SeccompProfile)
		apparmor = apparmorProfile(sc.ApparmorProfile)
	}

	if nonRoot == nil && user == nil && group == nil && fsGroup == nil && seccomp == nil {
		return nil
	}

	securityContext := &v1.PodSecurityContext{
		RunAsNonRoot:    nonRoot,
		RunAsUser:       user,
		RunAsGroup:      group,
		FSGroup:         fsGroup,
		SeccompProfile:  seccomp,
		AppArmorProfile: apparmor,
	}
	log.Trace().Msgf("pod security context that will be used: %v", securityContext)
	return securityContext
}

func seccompProfile(scp *SecProfile) *v1.SeccompProfile {
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

func apparmorProfile(scp *SecProfile) *v1.AppArmorProfile {
	if scp == nil || len(scp.Type) == 0 {
		return nil
	}
	log.Trace().Msgf("using AppArmor profile: %v", scp)

	apparmorProfile := &v1.AppArmorProfile{
		Type: v1.AppArmorProfileType(scp.Type),
	}
	if len(scp.LocalhostProfile) > 0 {
		apparmorProfile.LocalhostProfile = &scp.LocalhostProfile
	}

	return apparmorProfile
}

func containerSecurityContext(sc *SecurityContext, stepPrivileged bool) *v1.SecurityContext {
	if !stepPrivileged {
		return nil
	}

	privileged := false

	// if security context privileged is set explicitly
	if sc != nil && sc.Privileged != nil && *sc.Privileged {
		privileged = true
	}

	// if security context privileged is not set explicitly, but step is privileged
	if (sc == nil || sc.Privileged == nil) && stepPrivileged {
		privileged = true
	}

	if privileged {
		securityContext := &v1.SecurityContext{
			Privileged: newBool(true),
		}
		log.Trace().Msgf("container security context that will be used: %v", securityContext)
		return securityContext
	}

	return nil
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

func startPod(ctx context.Context, engine *kube, step *types.Step, options BackendOptions) (*v1.Pod, error) {
	podName, err := stepToPodName(step)
	if err != nil {
		return nil, err
	}
	engineConfig := engine.getConfig()
	pod, err := mkPod(step, engineConfig, podName, engine.goos, options)
	if err != nil {
		return nil, err
	}

	log.Trace().Msgf("creating pod: %s", pod.Name)
	return engine.client.CoreV1().Pods(engineConfig.Namespace).Create(ctx, pod, meta_v1.CreateOptions{})
}

func stopPod(ctx context.Context, engine *kube, step *types.Step, deleteOpts meta_v1.DeleteOptions) error {
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
