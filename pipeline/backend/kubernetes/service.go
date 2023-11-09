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

	"github.com/rs/zerolog/log"
	"go.woodpecker-ci.org/woodpecker/pipeline/backend/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Service(namespace, name string, ports []uint16, selector map[string]string) (*v1.Service, error) {
	log.Trace().Str("name", name).Interface("selector", selector).Interface("ports", ports).Msg("Creating service")

	var svcPorts []v1.ServicePort
	for _, port := range ports {
		svcPorts = append(svcPorts, v1.ServicePort{
			Port:       int32(port),
			TargetPort: intstr.IntOrString{IntVal: int32(port)},
		})
	}

	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeClusterIP,
			Selector: selector,
			Ports:    svcPorts,
		},
	}, nil
}

func ServiceName(step *types.Step) (string, error) {
	return dnsName(step.Name)
}

func StartService(ctx context.Context, engine *kube, step *types.Step) (*v1.Service, error) {
	name, err := ServiceName(step)
	if err != nil {
		return nil, err
	}
	podName, err := PodName(step)
	if err != nil {
		return nil, err
	}

	selector := map[string]string{
		StepLabel: podName,
	}

	svc, err := Service(engine.config.Namespace, name, step.Ports, selector)
	if err != nil {
		return nil, err
	}

	return engine.client.CoreV1().Services(engine.config.Namespace).Create(ctx, svc, metav1.CreateOptions{})
}

func StopService(ctx context.Context, engine *kube, step *types.Step, deleteOpts metav1.DeleteOptions) error {
	svcName, err := ServiceName(step)
	if err != nil {
		return err
	}
	log.Trace().Str("name", svcName).Msg("Deleting service")

	err = engine.client.CoreV1().Services(engine.config.Namespace).Delete(ctx, svcName, deleteOpts)
	if errors.IsNotFound(err) {
		// Don't abort on 404 errors from k8s, they most likely mean that the pod hasn't been created yet, usually because pipeline was canceled before running all steps.
		log.Trace().Err(err).Msgf("Unable to delete service %s", svcName)
		return nil
	}
	return err
}
