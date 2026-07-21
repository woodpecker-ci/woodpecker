// Copyright 2026 Woodpecker Authors
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

func TestRegistryValidate(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr error
	}{
		{name: "Docker Hub", address: "docker.io"},
		{name: "Docker Hub legacy domain", address: "index.docker.io"},
		{name: "domain", address: "registry.example.com"},
		{name: "domain with port", address: "registry.example.com:5000"},
		{name: "IPv4 with port", address: "10.0.1.32:5000"},
		{name: "IPv6 with port", address: "[2001:db8::1]:5000"},
		{name: "localhost with port", address: "localhost:5000"},
		{name: "empty", wantErr: errRegistryAddressInvalid},
		{name: "scheme", address: "http://registry.example.com", wantErr: errRegistryAddressInvalid},
		{name: "path", address: "registry.example.com/team", wantErr: errRegistryAddressInvalid},
		{name: "credentials", address: "user@registry.example.com", wantErr: errRegistryAddressInvalid},
		{name: "malformed port", address: "registry.example.com:port", wantErr: errRegistryAddressInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := Registry{
				Address:  test.address,
				Username: "user",
				Password: "password",
			}

			assert.ErrorIs(t, registry.Validate(), test.wantErr)
		})
	}
}

func TestRegistryValidateCredentials(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  error
	}{
		{name: "valid", username: "user", password: "password"},
		{name: "missing username", password: "password", wantErr: errRegistryUsernameInvalid},
		{name: "missing password", username: "user", wantErr: errRegistryPasswordInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := Registry{
				Address:  "registry.example.com",
				Username: test.username,
				Password: test.password,
			}

			assert.ErrorIs(t, registry.Validate(), test.wantErr)
		})
	}
}
