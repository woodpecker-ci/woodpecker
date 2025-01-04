// Copyright 2018 Drone.IO Inc.
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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretValidate(t *testing.T) {
	tests := []struct {
		s   Secret
		err bool
	}{
		{
			s: Secret{
				Name:   "secretname",
				Value:  "secretvalue",
				Events: []WebhookEvent{EventPush},
				Images: []string{"docker.io/library/mysql:latest", "alpine:latest", "localregistry.test:8443/mysql:latest", "localregistry.test:8443/library/mysql:latest", "docker.io/library/mysql", "alpine", "localregistry.test:8443/mysql", "localregistry.test:8443/library/mysql"},
			},
			err: false,
		},
		{
			s: Secret{
				Value:  "secretvalue",
				Events: []WebhookEvent{EventPush},
				Images: []string{"docker.io/library/mysql:latest", "alpine:latest", "localregistry.test:8443/mysql:latest", "localregistry.test:8443/library/mysql:latest", "docker.io/library/mysql", "alpine", "localregistry.test:8443/mysql", "localregistry.test:8443/library/mysql"},
			},
			err: true,
		},
		{
			s: Secret{
				Name:   "secretname",
				Events: []WebhookEvent{EventPush},
				Images: []string{"docker.io/library/mysql:latest", "alpine:latest", "localregistry.test:8443/mysql:latest", "localregistry.test:8443/library/mysql:latest", "docker.io/library/mysql", "alpine", "localregistry.test:8443/mysql", "localregistry.test:8443/library/mysql"},
			},
			err: true,
		},
		{
			s: Secret{
				Name:   "secretname",
				Value:  "secretvalue",
				Images: []string{"docker.io/library/mysql:latest", "alpine:latest", "localregistry.test:8443/mysql:latest", "localregistry.test:8443/library/mysql:latest", "docker.io/library/mysql", "alpine", "localregistry.test:8443/mysql", "localregistry.test:8443/library/mysql"},
			},
			err: true,
		},
		{
			s: Secret{
				Name:   "secretname",
				Value:  "secretvalue",
				Events: []WebhookEvent{EventPush},
				Images: []string{"wrong image:no"},
			},
			err: true,
		},
		{
			s: Secret{
				Name:   "secretname",
				Value:  "secretvalue",
				Events: []WebhookEvent{EventPush},
				Images: []string{"/library/mysql:latest", ":8443/mysql:latest", ":8443/library/mysql:latest", "/library/mysql", ":8443/mysql", ":8443/library/mysql"},
			},
			err: true,
		},
		{
			s: Secret{
				Name:   "secretname",
				Value:  "secretvalue",
				Events: []WebhookEvent{EventPush},
				Images: []string{"localregistry.test:/mysql:latest", "localregistry.test:/mysql"},
			},
			err: true,
		},
		{
			s: Secret{
				Name:   "secretname",
				Value:  "secretvalue",
				Events: []WebhookEvent{EventPush},
				Images: []string{"docker.io/library/mysql:", "alpine:", "localregistry.test:8443/mysql:", "localregistry.test:8443/library/mysql:"},
			},
			err: true,
		},
	}
	for i, tt := range tests {
		err := tt.s.Validate()
		if tt.err {
			assert.Errorf(t, err, "expected secret validation error on index %d", i)
		} else {
			assert.NoErrorf(t, err, "unexpected secret validation error on index %d", i)
		}
	}
}
