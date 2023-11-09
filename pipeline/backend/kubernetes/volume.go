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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PersistentVolumeClaim(namespace, name, storageClass, size string, storageRwx bool) (*v1.PersistentVolumeClaim, error) {
	_storageClass := &storageClass
	if storageClass == "" {
		_storageClass = nil
	}

	var accessMode v1.PersistentVolumeAccessMode

	if storageRwx {
		accessMode = v1.ReadWriteMany
	} else {
		accessMode = v1.ReadWriteOnce
	}

	volumeName, err := VolumeName(name)
	if err != nil {
		return nil, err
	}

	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      volumeName,
			Namespace: namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes:      []v1.PersistentVolumeAccessMode{accessMode},
			StorageClassName: _storageClass,
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(size),
				},
			},
		},
	}

	return pvc, nil
}

func VolumeName(name string) (string, error) {
	return dnsName(strings.Split(name, ":")[0])
}

func VolumeMountPath(name string) string {
	s := strings.Split(name, ":")
	if len(s) > 1 {
		return s[1]
	}
	return s[0]
}

func StartVolume(ctx context.Context, engine *kube, name string) (*v1.PersistentVolumeClaim, error) {
	pvc, err := PersistentVolumeClaim(engine.config.Namespace, name, engine.config.StorageClass, engine.config.VolumeSize, engine.config.StorageRwx)
	if err != nil {
		return nil, err
	}

	log.Trace().Msgf("Creating volume: %s", pvc.Name)
	return engine.client.CoreV1().PersistentVolumeClaims(engine.config.Namespace).Create(ctx, pvc, metav1.CreateOptions{})
}

func StopVolume(ctx context.Context, engine *kube, name string, deleteOpts metav1.DeleteOptions) error {
	pvcName, err := VolumeName(name)
	if err != nil {
		return err
	}
	log.Trace().Str("name", pvcName).Msg("Deleting volume")

	err = engine.client.CoreV1().PersistentVolumeClaims(engine.config.Namespace).Delete(ctx, pvcName, deleteOpts)
	if errors.IsNotFound(err) {
		// Don't abort on 404 errors from k8s, they most likely mean that the pod hasn't been created yet, usually because pipeline was canceled before running all steps.
		log.Trace().Err(err).Msgf("Unable to delete service %s", pvcName)
		return nil
	}
	return err
}
