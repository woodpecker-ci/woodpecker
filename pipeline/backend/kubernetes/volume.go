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
	"strings"

	"github.com/rs/zerolog/log"
	kube_core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	kube_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func mkPersistentVolumeClaim(config *config, name, namespace string) (*kube_core_v1.PersistentVolumeClaim, error) {
	_storageClass := &config.StorageClass
	if config.StorageClass == "" {
		_storageClass = nil
	}

	var accessMode kube_core_v1.PersistentVolumeAccessMode

	if config.StorageRwx {
		accessMode = kube_core_v1.ReadWriteMany
	} else {
		accessMode = kube_core_v1.ReadWriteOnce
	}

	volumeName, err := volumeName(name)
	if err != nil {
		return nil, err
	}

	pvc := &kube_core_v1.PersistentVolumeClaim{
		ObjectMeta: kube_meta_v1.ObjectMeta{
			Name:      volumeName,
			Namespace: namespace,
		},
		Spec: kube_core_v1.PersistentVolumeClaimSpec{
			AccessModes:      []kube_core_v1.PersistentVolumeAccessMode{accessMode},
			StorageClassName: _storageClass,
			Resources: kube_core_v1.VolumeResourceRequirements{
				Requests: kube_core_v1.ResourceList{
					kube_core_v1.ResourceStorage: resource.MustParse(config.VolumeSize),
				},
			},
		},
	}

	return pvc, nil
}

func volumeName(name string) (string, error) {
	return toDNSName(strings.Split(name, ":")[0])
}

func volumeMountPath(name string) string {
	s := strings.Split(name, ":")
	if len(s) > 1 {
		return s[1]
	}
	return s[0]
}

func startVolume(ctx context.Context, engine *kube, name, namespace string) (*kube_core_v1.PersistentVolumeClaim, error) {
	engineConfig := engine.getConfig()
	pvc, err := mkPersistentVolumeClaim(engineConfig, name, namespace)
	if err != nil {
		return nil, err
	}

	log.Trace().Msgf("creating volume: %s", pvc.Name)
	return engine.client.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, kube_meta_v1.CreateOptions{})
}

func stopVolume(ctx context.Context, engine *kube, name, namespace string, deleteOpts kube_meta_v1.DeleteOptions) error {
	pvcName, err := volumeName(name)
	if err != nil {
		return err
	}
	log.Trace().Str("name", pvcName).Msg("deleting volume")

	err = engine.client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvcName, deleteOpts)
	if errors.IsNotFound(err) {
		// Don't abort on 404 errors from k8s, they most likely mean that the pod hasn't been created yet, usually because pipeline was canceled before running all steps.
		log.Trace().Err(err).Msgf("unable to delete service %s", pvcName)
		return nil
	}
	return err
}

// useWorkflowVolumeForWorkspace points the workspace mount of every step at the
// workflow volume this backend creates. Callers may name the workspace volume
// differently from config.Volume, which left pods referencing a PVC that never
// existed (see https://github.com/woodpecker-ci/woodpecker/issues/7148). Local
// filesystem paths mounted inside the workspace are dropped, because Kubernetes
// cannot provide them and the workspace volume already covers that path.
func useWorkflowVolumeForWorkspace(conf *types.Config) error {
	workflowVolume, err := volumeName(conf.Volume)
	if err != nil {
		return err
	}

	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			volumes := make([]string, 0, len(step.Volumes))
			for _, volume := range step.Volumes {
				mountPath := volumeMountPath(volume)
				switch {
				case mountPath == step.WorkspaceBase:
					volumes = append(volumes, workflowVolume+":"+mountPath)
				case isLocalPath(volume) && isBelowPath(mountPath, step.WorkspaceBase):
					// Covered by the workspace volume mounted at the workspace base.
				default:
					volumes = append(volumes, volume)
				}
			}
			step.Volumes = volumes
		}
	}

	return nil
}

// isLocalPath reports whether a volume is backed by a local filesystem path
// instead of an existing PVC.
func isLocalPath(volume string) bool {
	return strings.HasPrefix(strings.Split(volume, ":")[0], "/")
}

// isBelowPath reports whether path is located below base.
func isBelowPath(path, base string) bool {
	return base != "" && strings.HasPrefix(path, strings.TrimSuffix(base, "/")+"/")
}
