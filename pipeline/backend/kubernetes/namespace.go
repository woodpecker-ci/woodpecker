// Copyright 2025 Woodpecker Authors
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
	"context"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sNamespaceClient interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Namespace, error)
	Create(ctx context.Context, namespace *v1.Namespace, opts metav1.CreateOptions) (*v1.Namespace, error)
}

func mkNamespace(ctx context.Context, client K8sNamespaceClient, namespace string) error {
	_, err := client.Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		log.Trace().Str("namespace", namespace).Msg("Kubernetes namespace already exists")
		return nil
	}

	if !errors.IsNotFound(err) {
		log.Trace().Err(err).Str("namespace", namespace).Msg("failed to check Kubernetes namespace existence")
		return err
	}

	log.Trace().Str("namespace", namespace).Msg("creating Kubernetes namespace")

	_, err = client.Create(ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: namespace},
	}, metav1.CreateOptions{})
	if err != nil {
		log.Error().Err(err).Str("namespace", namespace).Msg("failed to create Kubernetes namespace")
		return err
	}

	log.Trace().Str("namespace", namespace).Msg("Kubernetes namespace created successfully")
	return nil
}
