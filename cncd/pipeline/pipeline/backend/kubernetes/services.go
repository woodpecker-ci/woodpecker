package kubernetes

import (
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Service(namespace, name, podName string, ports []string) (*v1.Service, error) {
	var svcPorts []v1.ServicePort
	for _, p := range ports {
		i, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("Could not parse service port %s as integer", p)
		}
		svcPorts = append(svcPorts, v1.ServicePort{
			Port:       int32(i),
			TargetPort: intstr.IntOrString{IntVal: int32(i)},
		})
	}

	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dnsName("dr-" + name),
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"step": podName,
			},
			Ports: svcPorts,
		},
	}, nil
}
