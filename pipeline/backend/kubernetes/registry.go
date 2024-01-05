// Copyright 2023 Woodpecker Authors
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
	"encoding/json"
	"github.com/rs/zerolog/log"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	k8s "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type ConfigAuths struct {
	Auths map[string]ConfigAuth `json:"auths,omitempty"`
}

type ConfigAuth struct {
	Auth string `json:"auth,omitempty"`
}

func startRegistriesAuth(ctx context.Context, engine *kube, registries []*types.Registry, taskUUID string) (*k8s.Secret, error) {
	regNames := make([]string, len(registries))
	for i, reg := range registries {
		regNames[i] = reg.Hostname
	}
	log.Debug().
		Str("registries", strings.Join(regNames, ",")).
		Msg("creating images pull secret")

	authsJsonBytes, err := dockerAuths(registries)
	if err != nil {
		return nil, err
	}

	pullSecret, err := mkPullSecret(engine.config.Namespace, taskUUID, authsJsonBytes)
	if err != nil {
		return nil, err
	}

	return engine.client.CoreV1().Secrets(engine.config.Namespace).Create(ctx, pullSecret, k8smeta.CreateOptions{})
}

func stopRegistriesAuth(ctx context.Context, engine *kube, secretName string, deleteOpts k8smeta.DeleteOptions) error {
	log.Trace().Str("name", secretName).Msg("Deleting pull secret")

	err := engine.client.CoreV1().Secrets(engine.config.Namespace).Delete(ctx, secretName, deleteOpts)
	if errors.IsNotFound(err) {
		// Don't abort on 404 errors from k8s, they most likely mean that the pod hasn't been created yet, usually because pipeline was canceled before running all steps.
		log.Trace().Err(err).Msgf("Unable to delete pull secret %s", secretName)
		return nil
	}
	return err
}

func dockerAuths(registries []*types.Registry) ([]byte, error) {
	auths, err := configAuths(registries)
	if err != nil {
		return nil, err
	}
	return json.Marshal(auths)
}

func configAuths(registries []*types.Registry) (*ConfigAuths, error) {
	auths := make(map[string]ConfigAuth, len(registries))
	for _, reg := range registries {
		auth, err := configAuth(reg)
		if err != nil {
			return nil, err
		}
		auths[reg.Hostname] = auth
	}
	return &ConfigAuths{Auths: auths}, nil
}

func configAuth(registry *types.Registry) (ConfigAuth, error) {
	authB64, err := types.Auth{
		Username: registry.Username,
		Password: registry.Password,
		Email:    registry.Email,
	}.EncodeToBase64()
	return ConfigAuth{Auth: authB64}, err
}

func mkPullSecret(namespace, name string, authsJsonBytes []byte) (*k8s.Secret, error) {
	meta := k8smeta.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}

	data := map[string][]byte{k8s.DockerConfigJsonKey: authsJsonBytes}
	return &k8s.Secret{
		ObjectMeta: meta,
		Data:       data,
		Type:       k8s.SecretTypeDockerConfigJson,
	}, nil
}
