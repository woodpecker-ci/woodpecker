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
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mkPersistentVolumeClaim(config *config, name string) (*v1.PersistentVolumeClaim, error) {
	_storageClass := &config.StorageClass
	if config.StorageClass == "" {
		_storageClass = nil
	}

	var accessMode v1.PersistentVolumeAccessMode

	if config.StorageRwx {
		accessMode = v1.ReadWriteMany
	} else {
		accessMode = v1.ReadWriteOnce
	}

	volumeName, err := volumeName(name)
	if err != nil {
		return nil, err
	}

	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      volumeName,
			Namespace: config.Namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes:      []v1.PersistentVolumeAccessMode{accessMode},
			StorageClassName: _storageClass,
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(config.VolumeSize),
				},
			},
		},
	}

	return pvc, nil
}

func volumeName(name string) (string, error) {
	return dnsName(strings.Split(name, ":")[0])
}

func volumeMountPath(name string) string {
	s := strings.Split(name, ":")
	if len(s) > 1 {
		return s[1]
	}
	return s[0]
}

func startVolume(ctx context.Context, engine *kube, name string) (*v1.PersistentVolumeClaim, error) {
	engineConfig := engine.getConfig()
	pvc, err := mkPersistentVolumeClaim(engineConfig, name)
	if err != nil {
		return nil, err
	}

	log.Trace().Msgf("creating volume: %s", pvc.Name)
	return engine.client.CoreV1().PersistentVolumeClaims(engineConfig.Namespace).Create(ctx, pvc, meta_v1.CreateOptions{})
}

func stopVolume(ctx context.Context, engine *kube, name string, deleteOpts meta_v1.DeleteOptions) error {
	pvcName, err := volumeName(name)
	if err != nil {
		return err
	}
	log.Trace().Str("name", pvcName).Msg("deleting volume")

	err = engine.client.CoreV1().PersistentVolumeClaims(engine.config.Namespace).Delete(ctx, pvcName, deleteOpts)
	if errors.IsNotFound(err) {
		// Don't abort on 404 errors from k8s, they most likely mean that the pod hasn't been created yet, usually because pipeline was canceled before running all steps.
		log.Trace().Err(err).Msgf("unable to delete service %s", pvcName)
		return nil
	}
	return err
}
