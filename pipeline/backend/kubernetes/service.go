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
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	int_str "k8s.io/apimachinery/pkg/util/intstr"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

const (
	ServiceLabel  = "service"
	servicePrefix = "wp-svc-"
)

func mkService(step *types.Step, config *config, workflowName string) (*v1.Service, error) {
	name, err := serviceName(step, workflowName)
	if err != nil {
		return nil, err
	}

	selector := map[string]string{
		ServiceLabel: name,
	}

	var svcPorts []v1.ServicePort
	for _, port := range step.Ports {
		svcPorts = append(svcPorts, servicePort(port))
	}

	log.Trace().Str("name", name).Interface("selector", selector).Interface("ports", svcPorts).Msg("creating service")
	return &v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: config.Namespace,
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeClusterIP,
			Selector: selector,
			Ports:    svcPorts,
		},
	}, nil
}

func serviceName(step *types.Step, workflowName string) (string, error) {
	return dnsName(servicePrefix + workflowName + "-" + step.Name + "-" + step.UUID[len(step.UUID)-5:])
}

func servicePort(port types.Port) v1.ServicePort {
	portNumber := int32(port.Number)
	portProtocol := strings.ToUpper(port.Protocol)
	return v1.ServicePort{
		Name:       fmt.Sprintf("port-%d", portNumber),
		Port:       portNumber,
		Protocol:   v1.Protocol(portProtocol),
		TargetPort: int_str.IntOrString{IntVal: portNumber},
	}
}

func startService(ctx context.Context, engine *kube, step *types.Step, workflowName string) (*v1.Service, error) {
	engineConfig := engine.getConfig()
	svc, err := mkService(step, engineConfig, workflowName)
	if err != nil {
		return nil, err
	}

	log.Trace().Str("name", svc.Name).Interface("selector", svc.Spec.Selector).Interface("ports", svc.Spec.Ports).Msg("creating service")
	return engine.client.CoreV1().Services(engineConfig.Namespace).Create(ctx, svc, meta_v1.CreateOptions{})
}

func stopService(ctx context.Context, engine *kube, step *types.Step, deleteOpts meta_v1.DeleteOptions, workflowName string) error {
	svcName, err := serviceName(step, workflowName)
	if err != nil {
		return err
	}
	log.Trace().Str("name", svcName).Msg("deleting service")

	err = engine.client.CoreV1().Services(engine.config.Namespace).Delete(ctx, svcName, deleteOpts)
	if errors.IsNotFound(err) {
		// Don't abort on 404 errors from k8s, they most likely mean that the pod hasn't been created yet, usually because pipeline was canceled before running all steps.
		log.Trace().Err(err).Msgf("unable to delete service %s", svcName)
		return nil
	}
	return err
}
