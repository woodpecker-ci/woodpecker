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
