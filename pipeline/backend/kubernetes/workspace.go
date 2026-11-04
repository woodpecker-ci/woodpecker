// Copyright 2026 Woodpecker Authors
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
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	kube_core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	kube_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// workspaceOverlay is how a step with the overlay workspace mode gets its own copy of the workspace.
type workspaceOverlay int

const (
	// Mount a volume cloned from the workspace volume by the CSI driver.
	overlayClone workspaceOverlay = iota + 1
	// Copy the workspace into an emptyDir volume with an init container.
	overlayCopy
)

const (
	cloneProbePrefix      = "wp-clone-probe-"
	cloneProbeLabel       = "woodpecker-ci.org/clone-probe"
	cloneProbeUser        = 65534
	overlayInitName       = "init-workspace-overlay"
	overlaySourceSuffix   = "-source"
	overlaySourcePath     = "/workspace-source"
	overlayTargetPath     = "/workspace-target"
	eventReasonProvFailed = "ProvisioningFailed"
)

var (
	cloneProbeTimeout  = 2 * time.Minute
	cloneProbeInterval = 2 * time.Second
)

// cloneProbeCache keeps per storage class whether the CSI driver can clone the workspace volume.
type cloneProbeCache struct {
	mu      sync.Mutex
	results map[string]bool
}

// workspaceOverlayMode returns how an overlay step gets its copy of the workspace. The first
// time a storage class needs it, a short-lived clone of the workspace claim shows whether the
// CSI driver supports cloning; the result is kept until the agent restarts.
func (e *kube) workspaceOverlayMode(ctx context.Context, namespace, workspaceClaim string) (workspaceOverlay, error) {
	cfg := e.getConfig()

	e.cloneProbe.mu.Lock()
	defer e.cloneProbe.mu.Unlock()

	if supported, ok := e.cloneProbe.results[cfg.StorageClass]; ok {
		return overlayFor(supported), nil
	}

	source, err := e.client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, workspaceClaim, kube_meta_v1.GetOptions{})
	if err != nil {
		return 0, err
	}
	if source.Status.Phase != kube_core_v1.ClaimBound {
		// only a bound claim can be cloned, this says nothing about the driver
		return overlayCopy, nil
	}

	supported, err := probeClone(ctx, e.client, cfg, namespace, workspaceClaim)
	if err != nil {
		return 0, err
	}
	if e.cloneProbe.results == nil {
		e.cloneProbe.results = make(map[string]bool)
	}
	e.cloneProbe.results[cfg.StorageClass] = supported
	return overlayFor(supported), nil
}

func overlayFor(cloneSupported bool) workspaceOverlay {
	if cloneSupported {
		return overlayClone
	}
	return overlayCopy
}

// probeClone clones the workspace claim with a pod consuming the clone, so storage classes
// that wait for the first consumer provision it too, and reports whether the clone got bound.
func probeClone(ctx context.Context, client kubernetes.Interface, cfg *config, namespace, workspaceClaim string) (bool, error) {
	name := cloneProbePrefix + rand.String(10) //nolint:mnd

	pvc, err := clonedPersistentVolumeClaim(cfg, name, namespace, workspaceClaim)
	if err != nil {
		return false, err
	}
	pvc.Labels = map[string]string{cloneProbeLabel: "true"}
	if _, err := client.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, kube_meta_v1.CreateOptions{}); err != nil {
		return false, err
	}
	defer removeCloneProbe(context.WithoutCancel(ctx), client, namespace, pvc.Name)

	if _, err := client.CoreV1().Pods(namespace).Create(ctx, cloneProbePod(cfg, pvc.Name, namespace), kube_meta_v1.CreateOptions{}); err != nil {
		return false, err
	}

	timeout := time.NewTimer(cloneProbeTimeout)
	defer timeout.Stop()
	ticker := time.NewTicker(cloneProbeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-timeout.C:
			log.Warn().Str("storageClass", cfg.StorageClass).Msg("cloning the workspace volume did not finish in time, overlay steps copy the workspace")
			return false, nil
		case <-ticker.C:
		}

		clone, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvc.Name, kube_meta_v1.GetOptions{})
		if err != nil {
			return false, err
		}
		if clone.Status.Phase == kube_core_v1.ClaimBound {
			log.Debug().Str("storageClass", cfg.StorageClass).Msg("the CSI driver can clone the workspace volume, overlay steps use clones")
			return true, nil
		}

		// the agent might not be allowed to read events, then only the timeout decides
		events, err := client.CoreV1().Events(namespace).List(ctx, kube_meta_v1.ListOptions{
			FieldSelector: "involvedObject.name=" + pvc.Name,
		})
		if err != nil {
			continue
		}
		for _, event := range events.Items {
			if event.InvolvedObject.Name == pvc.Name && event.Reason == eventReasonProvFailed {
				log.Info().Str("storageClass", cfg.StorageClass).Str("reason", event.Message).Msg("the workspace volume can not be cloned, overlay steps copy the workspace")
				return false, nil
			}
		}
	}
}

func clonedPersistentVolumeClaim(cfg *config, name, namespace, workspaceClaim string) (*kube_core_v1.PersistentVolumeClaim, error) {
	pvc, err := mkPersistentVolumeClaim(cfg, name, namespace)
	if err != nil {
		return nil, err
	}
	pvc.Spec.DataSource = &kube_core_v1.TypedLocalObjectReference{
		Kind: "PersistentVolumeClaim",
		Name: workspaceClaim,
	}
	return pvc, nil
}

func cloneProbePod(cfg *config, claim, namespace string) *kube_core_v1.Pod {
	return &kube_core_v1.Pod{
		Name:      claim,
		Namespace: namespace,
		Labels:    map[string]string{cloneProbeLabel: "true"},
		Spec: kube_core_v1.PodSpec{
			RestartPolicy:     kube_core_v1.RestartPolicyNever,
			PriorityClassName: cfg.PriorityClassName,
			NodeSelector:      cfg.PodNodeSelector,
			Tolerations:       tolerations(cfg.PodTolerations),
			Affinity:          cfg.PodAffinity,
			ImagePullSecrets:  secretsReferences(cfg.ImagePullSecretNames),
			Containers: []kube_core_v1.Container{{
				Name:         "probe",
				Image:        cfg.PermissionInitImage,
				Command:      []string{"true"},
				VolumeMounts: []kube_core_v1.VolumeMount{{Name: claim, MountPath: "/probe"}},
				SecurityContext: &kube_core_v1.SecurityContext{
					RunAsNonRoot:             newBool(true),
					RunAsUser:                newInt64(cloneProbeUser),
					AllowPrivilegeEscalation: newBool(false),
					Capabilities:             &kube_core_v1.Capabilities{Drop: []kube_core_v1.Capability{"ALL"}},
					SeccompProfile:           &kube_core_v1.SeccompProfile{Type: kube_core_v1.SeccompProfileTypeRuntimeDefault},
				},
			}},
			Volumes: []kube_core_v1.Volume{pvcVolume(claim)},
		},
	}
}

func removeCloneProbe(ctx context.Context, client kubernetes.Interface, namespace, name string) {
	ctx, cancel := context.WithTimeout(ctx, cloneProbeTimeout)
	defer cancel()
	if err := client.CoreV1().Pods(namespace).Delete(ctx, name, kube_meta_v1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		log.Error().Err(err).Str("pod", name).Msg("could not delete clone probe pod")
	}
	if err := client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, kube_meta_v1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		log.Error().Err(err).Str("pvc", name).Msg("could not delete clone probe volume claim")
	}
}

// workspaceClaim returns the claim the step container mounts at the workspace base, if any.
func workspaceClaim(pod *kube_core_v1.Pod, step *types.Step) (volumeIndex int, claim string, found bool) {
	if len(pod.Spec.Containers) == 0 {
		return 0, "", false
	}
	for _, mount := range pod.Spec.Containers[0].VolumeMounts {
		if mount.MountPath != step.WorkspaceBase {
			continue
		}
		for i, volume := range pod.Spec.Volumes {
			if volume.Name == mount.Name && volume.PersistentVolumeClaim != nil {
				return i, volume.PersistentVolumeClaim.ClaimName, true
			}
		}
	}
	return 0, "", false
}

// applyWorkspaceOverlay gives the step its own copy of the workspace, which is deleted
// with the pod. It reports false if the step does not mount the workspace.
func applyWorkspaceOverlay(pod *kube_core_v1.Pod, step *types.Step, cfg *config, overlay workspaceOverlay) bool {
	index, claim, found := workspaceClaim(pod, step)
	if !found {
		return false
	}
	volume := &pod.Spec.Volumes[index]

	switch overlay {
	case overlayClone:
		clone, err := clonedPersistentVolumeClaim(cfg, claim, pod.Namespace, claim)
		if err != nil {
			return false
		}
		volume.VolumeSource = kube_core_v1.VolumeSource{
			Ephemeral: &kube_core_v1.EphemeralVolumeSource{
				VolumeClaimTemplate: &kube_core_v1.PersistentVolumeClaimTemplate{Spec: clone.Spec},
			},
		}
	case overlayCopy:
		sourceName := volume.Name + overlaySourceSuffix
		volume.VolumeSource = kube_core_v1.VolumeSource{EmptyDir: &kube_core_v1.EmptyDirVolumeSource{}}
		pod.Spec.Volumes = append(pod.Spec.Volumes, kube_core_v1.Volume{
			Name:                  sourceName,
			PersistentVolumeClaim: &kube_core_v1.PersistentVolumeClaimVolumeSource{ClaimName: claim, ReadOnly: true},
		})
		// the init container runs as the user of the pod, so the step owns the copy
		initContainer := kube_core_v1.Container{
			Name:    overlayInitName,
			Image:   cfg.PermissionInitImage,
			Command: []string{"cp", "-a", overlaySourcePath + "/.", overlayTargetPath + "/"},
			VolumeMounts: []kube_core_v1.VolumeMount{
				{Name: sourceName, MountPath: overlaySourcePath, ReadOnly: true},
				{Name: pod.Spec.Volumes[index].Name, MountPath: overlayTargetPath},
			},
			SecurityContext: &kube_core_v1.SecurityContext{
				AllowPrivilegeEscalation: newBool(false),
				Capabilities:             &kube_core_v1.Capabilities{Drop: []kube_core_v1.Capability{"ALL"}},
			},
		}
		pod.Spec.InitContainers = append([]kube_core_v1.Container{initContainer}, pod.Spec.InitContainers...)
	default:
		return false
	}
	return true
}
