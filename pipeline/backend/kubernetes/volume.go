// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubernetes

import (
	"strings"

	v1 "k8s.io/api/core/v1"
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

	volumeName, err := dnsName(strings.Split(name, ":")[0])
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
