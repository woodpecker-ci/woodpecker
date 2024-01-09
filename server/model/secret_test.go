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

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"
)

func TestSecretMatch(t *testing.T) {
	tcl := []*struct {
		name   string
		secret Secret
		event  WebhookEvent
		match  bool
	}{
		{
			name:   "should match event",
			secret: Secret{Events: []WebhookEvent{"pull_request"}},
			event:  EventPull,
			match:  true,
		},
		{
			name:   "should not match event",
			secret: Secret{Events: []WebhookEvent{"pull_request"}},
			event:  EventPush,
			match:  false,
		},
		{
			name:   "should match when no event filters defined",
			secret: Secret{},
			event:  EventPull,
			match:  true,
		},
		{
			name:   "pull close should match pull",
			secret: Secret{Events: []WebhookEvent{"pull_request"}},
			event:  EventPullClosed,
			match:  true,
		},
	}

	for _, tc := range tcl {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.match, tc.secret.Match(tc.event))
		})
	}
}

func TestSecretValidate(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("Secret", func() {
		g.It("should pass validation", func() {
			secret := Secret{
				Name:   "secretname",
				Value:  "secretvalue",
				Events: []WebhookEvent{EventPush},
				Images: []string{"docker.io/library/mysql:latest", "alpine:latest", "localregistry.test:8443/mysql:latest", "localregistry.test:8443/library/mysql:latest", "docker.io/library/mysql", "alpine", "localregistry.test:8443/mysql", "localregistry.test:8443/library/mysql"},
			}
			err := secret.Validate()
			g.Assert(err).IsNil()
		})
		g.Describe("should fail validation", func() {
			g.It("when no name", func() {
				secret := Secret{
					Value:  "secretvalue",
					Events: []WebhookEvent{EventPush},
					Images: []string{"docker.io/library/mysql:latest", "alpine:latest", "localregistry.test:8443/mysql:latest", "localregistry.test:8443/library/mysql:latest", "docker.io/library/mysql", "alpine", "localregistry.test:8443/mysql", "localregistry.test:8443/library/mysql"},
				}
				err := secret.Validate()
				g.Assert(err).IsNotNil()
			})
			g.It("when no value", func() {
				secret := Secret{
					Name:   "secretname",
					Events: []WebhookEvent{EventPush},
					Images: []string{"docker.io/library/mysql:latest", "alpine:latest", "localregistry.test:8443/mysql:latest", "localregistry.test:8443/library/mysql:latest", "docker.io/library/mysql", "alpine", "localregistry.test:8443/mysql", "localregistry.test:8443/library/mysql"},
				}
				err := secret.Validate()
				g.Assert(err).IsNotNil()
			})
			g.It("when no events", func() {
				secret := Secret{
					Name:   "secretname",
					Value:  "secretvalue",
					Images: []string{"docker.io/library/mysql:latest", "alpine:latest", "localregistry.test:8443/mysql:latest", "localregistry.test:8443/library/mysql:latest", "docker.io/library/mysql", "alpine", "localregistry.test:8443/mysql", "localregistry.test:8443/library/mysql"},
				}
				err := secret.Validate()
				g.Assert(err).IsNotNil()
			})
			g.It("wrong image: no value", func() {
				secret := Secret{
					Name:   "secretname",
					Value:  "secretvalue",
					Events: []WebhookEvent{EventPush},
					Images: []string{"wrong image:no"},
				}
				err := secret.Validate()
				g.Assert(err).IsNotNil()
			})
			g.It("wrong image: no hostname", func() {
				secret := Secret{
					Name:   "secretname",
					Value:  "secretvalue",
					Events: []WebhookEvent{EventPush},
					Images: []string{"/library/mysql:latest", ":8443/mysql:latest", ":8443/library/mysql:latest", "/library/mysql", ":8443/mysql", ":8443/library/mysql"},
				}
				err := secret.Validate()
				g.Assert(err).IsNotNil()
			})
			g.It("wrong image: no port number", func() {
				secret := Secret{
					Name:   "secretname",
					Value:  "secretvalue",
					Events: []WebhookEvent{EventPush},
					Images: []string{"localregistry.test:/mysql:latest", "localregistry.test:/mysql"},
				}
				err := secret.Validate()
				g.Assert(err).IsNotNil()
			})
			g.It("wrong image: no tag name", func() {
				secret := Secret{
					Name:   "secretname",
					Value:  "secretvalue",
					Events: []WebhookEvent{EventPush},
					Images: []string{"docker.io/library/mysql:", "alpine:", "localregistry.test:8443/mysql:", "localregistry.test:8443/library/mysql:"},
				}
				err := secret.Validate()
				g.Assert(err).IsNotNil()
			})
		})
	})
}
