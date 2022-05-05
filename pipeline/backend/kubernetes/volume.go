package kubernetes

import (
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PersistentVolumeClaim(namespace, name string, storageClass string, size string) *v1.PersistentVolumeClaim {
	_storageClass := &storageClass
	if storageClass == "" {
		_storageClass = nil
	}

	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      volumeName(name),
			Namespace: namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			StorageClassName: _storageClass,
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(size),
				},
			},
		},
	}
}

func volumeName(i string) string {
	return dnsName(strings.Split(i, ":")[0])
}
