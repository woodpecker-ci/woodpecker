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
	ServiceLabel          = "service"
	HeadlessServicePrefix = "wp-hsvc-"
	ServicePrefix         = "wp-svc-"
)

func mkService(step *types.Step, config *config) (*v1.Service, error) {
	name, err := serviceName(step)
	if err != nil {
		return nil, err
	}

	selector := map[string]string{
		ServiceLabel: name,
	}

	if len(step.Ports) == 0 {
		return nil, fmt.Errorf("kubernetes backend requires explicitly exposed ports for service steps, add 'ports' configuration to step '%s'", step.Name)
	}

	var svcPorts []v1.ServicePort
	for _, port := range step.Ports {
		svcPorts = append(svcPorts, servicePort(port))
	}

	log.Trace().Str("name", name).Interface("selector", selector).Interface("ports", svcPorts).Msg("creating service")
	return &v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: config.GetNamespace(step.OrgID),
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeClusterIP,
			Selector: selector,
			Ports:    svcPorts,
		},
	}, nil
}

func mkHeadlessService(namespace, taskUUID string) (*v1.Service, error) {
	selector := map[string]string{
		TaskUUIDLabel: taskUUID,
	}

	name, err := subdomain(taskUUID)
	if err != nil {
		return nil, err
	}

	log.Trace().Str("name", name).Interface("selector", selector).Msg("creating headless service")
	return &v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: "None",
			Selector:  selector,
		},
	}, nil
}

func serviceName(step *types.Step) (string, error) {
	return dnsName(ServicePrefix + step.UUID + "-" + step.Name)
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

func startService(ctx context.Context, engine *kube, step *types.Step) (*v1.Service, error) {
	engineConfig := engine.getConfig()
	svc, err := mkService(step, engineConfig)
	if err != nil {
		return nil, err
	}

	log.Trace().Str("name", svc.Name).Interface("selector", svc.Spec.Selector).Interface("ports", svc.Spec.Ports).Msg("creating service")
	return engine.client.CoreV1().Services(engineConfig.GetNamespace(step.OrgID)).Create(ctx, svc, meta_v1.CreateOptions{})
}

func stopService(ctx context.Context, engine *kube, step *types.Step, deleteOpts meta_v1.DeleteOptions) error {
	svcName, err := serviceName(step)
	if err != nil {
		return err
	}
	log.Trace().Str("name", svcName).Msg("deleting service")

	err = engine.client.CoreV1().Services(engine.config.GetNamespace(step.OrgID)).Delete(ctx, svcName, deleteOpts)
	if errors.IsNotFound(err) {
		// Don't abort on 404 errors from k8s, they most likely mean that the pod hasn't been created yet, usually because pipeline was canceled before running all steps.
		log.Trace().Err(err).Msgf("unable to delete service %s", svcName)
		return nil
	}
	return err
}

func subdomain(taskUUID string) (string, error) {
	return dnsName(HeadlessServicePrefix + taskUUID)
}

func startHeadlessService(ctx context.Context, engine *kube, namespace, taskUUID string) (*v1.Service, error) {
	svc, err := mkHeadlessService(namespace, taskUUID)
	if err != nil {
		return nil, err
	}

	log.Trace().Str("name", svc.Name).Interface("selector", svc.Spec.Selector).Msg("creating headless service")
	return engine.client.CoreV1().Services(namespace).Create(ctx, svc, meta_v1.CreateOptions{})
}

func stopHeadlessService(ctx context.Context, engine *kube, namespace, taskUUID string) error {
	name, err := subdomain(taskUUID)
	if err != nil {
		return err
	}

	log.Trace().Str("name", name).Msg("deleting headless service")

	err = engine.client.CoreV1().Services(namespace).Delete(ctx, name, defaultDeleteOptions)
	if errors.IsNotFound(err) {
		// Don't abort on 404 errors from k8s, they most likely mean that the pod hasn't been created yet, usually because pipeline was canceled before running all steps.
		log.Trace().Err(err).Msgf("unable to delete headless service %s", name)
		return nil
	}
	return err
}
