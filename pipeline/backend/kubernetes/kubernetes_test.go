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
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, false, engine.config.StorageRwx)
	assert.Equal(t, 1, len(engine.config.PodLabels))
	assert.Equal(t, 1, len(engine.config.PodAnnotations))
	assert.Equal(t, 1, len(engine.config.ImagePullSecretNames))
	assert.Equal(t, false, engine.config.SecurityContext.RunAsNonRoot)
}
