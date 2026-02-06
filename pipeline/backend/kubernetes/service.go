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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

const (
	ServiceLabel          = "service"
	HeadlessServicePrefix = "wp-hsvc-"
	ServicePrefix         = "wp-svc-"
)

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

func isService(step *types.Step) bool {
	return step.Type == types.StepTypeService || (step.Detached && dnsPattern.FindStringIndex(step.Name) != nil)
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
