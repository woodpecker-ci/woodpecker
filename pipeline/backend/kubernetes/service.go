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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Service(namespace, name, podName string, ports []uint16) (*v1.Service, error) {
	var svcPorts []v1.ServicePort
	for _, port := range ports {
		svcPorts = append(svcPorts, v1.ServicePort{
			Port:       int32(port),
			TargetPort: intstr.IntOrString{IntVal: int32(port)},
		})
	}

	dnsName, err := dnsName(name)
	if err != nil {
		return nil, err
	}

	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dnsName,
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
