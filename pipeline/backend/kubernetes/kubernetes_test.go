// Copyright 2024 Woodpecker Authors
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
	"testing"

	"github.com/stretchr/testify/assert"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestGettingConfig(t *testing.T) {
	engine := kube{
		config: &config{
			Namespace:            "default",
			StorageClass:         "hdd",
			VolumeSize:           "1G",
			StorageRwx:           false,
			PodLabels:            map[string]string{"l1": "v1"},
			PodAnnotations:       map[string]string{"a1": "v1"},
			ImagePullSecretNames: []string{"regcred"},
			SecurityContext:      SecurityContextConfig{RunAsNonRoot: false},
		},
	}
	config := engine.getConfig()
	config.Namespace = "wp"
	config.StorageClass = "ssd"
	config.StorageRwx = true
	config.PodLabels = nil
	config.PodAnnotations["a2"] = "v2"
	config.ImagePullSecretNames = append(config.ImagePullSecretNames, "docker.io")
	config.SecurityContext.RunAsNonRoot = true

	assert.Equal(t, "default", engine.config.Namespace)
	assert.Equal(t, "hdd", engine.config.StorageClass)
	assert.Equal(t, "1G", engine.config.VolumeSize)
	assert.False(t, engine.config.StorageRwx)
	assert.Len(t, engine.config.PodLabels, 1)
	assert.Len(t, engine.config.PodAnnotations, 1)
	assert.Len(t, engine.config.ImagePullSecretNames, 1)
	assert.False(t, engine.config.SecurityContext.RunAsNonRoot)
}

func TestSetupWorkflow(t *testing.T) {
	namespace := "foo"
	volumeName := "volume-name"
	volumePath := volumeName + ":/woodpecker"
	networkName := "test-network"
	taskUUID := "11301"

	engine := kube{
		config: &config{
			Namespace:            namespace,
			StorageClass:         "hdd",
			VolumeSize:           "1G",
			StorageRwx:           false,
			PodLabels:            map[string]string{"l1": "v1"},
			PodAnnotations:       map[string]string{"a1": "v1"},
			ImagePullSecretNames: []string{"regcred"},
			SecurityContext:      SecurityContextConfig{RunAsNonRoot: false},
		},
		client: fake.NewClientset(),
	}

	serviceWithPorts := types.Step{
		OrgID:    42,
		Name:     "service",
		UUID:     "123",
		Type:     types.StepTypeService,
		Volumes:  []string{volumePath},
		Networks: []types.Conn{{Name: networkName, Aliases: []string{"alias"}}},
		Ports: []types.Port{
			{Number: 8080, Protocol: "tcp"},
		},
	}

	conf := &types.Config{
		Volume:  volumePath,
		Network: networkName,
		Stages: []*types.Stage{
			{
				Steps: []*types.Step{
					&serviceWithPorts,
					{
						OrgID:    42,
						UUID:     "234",
						Name:     "service2",
						Type:     types.StepTypeService,
						Volumes:  []string{volumePath},
						Networks: []types.Conn{{Name: networkName, Aliases: []string{"alias"}}},
					},
				},
			},
			{
				Steps: []*types.Step{
					{
						OrgID:    42,
						UUID:     "456",
						Name:     "step-1",
						Volumes:  []string{volumePath},
						Networks: []types.Conn{{Name: networkName, Aliases: []string{"alias"}}},
					},
				},
			},
		},
	}

	err := engine.SetupWorkflow(context.Background(), conf, taskUUID)
	assert.NoError(t, err, "SetupWorkflow should not error with minimal config and fake client")

	_, err = engine.client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), "volume-name", meta_v1.GetOptions{})
	assert.NoError(t, err, "persistent volume should be created during workflow setup")

	_, err = engine.client.CoreV1().Services(namespace).Get(context.Background(), "wp-hsvc-"+taskUUID, meta_v1.GetOptions{})
	assert.NoError(t, err, "headless service should be created during workflow setup")
}
