package kubernetes

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Service(namespace, name, podName string, ports []int) *v1.Service {
	// We don't need a service, if we don't have ports
	if len(ports) == 0 {
		return nil
	}

	var svcPorts []v1.ServicePort
	for _, p := range ports {
		svcPorts = append(svcPorts, v1.ServicePort{
			Port:       int32(p),
			TargetPort: intstr.IntOrString{IntVal: int32(p)},
		})
	}

	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dnsName(name),
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"step": podName,
			},
			Ports: svcPorts,
		},
	}
}
