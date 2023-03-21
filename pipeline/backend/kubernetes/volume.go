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
