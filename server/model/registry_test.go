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

func TestRegistryValidateAddress(t *testing.T) {
	for _, address := range []string{
		"docker.io",
		"ghcr.io",
		"registry.example.com:5000",
		"10.0.1.32:5000",
		"localhost:5000",
		"https://registry.example.com",
	} {
		r := Registry{Address: address, Username: "user", Password: "pass"}
		assert.NoError(t, r.Validate(), address)
	}

	for _, address := range []string{"", "registry.example.com:port", "10.0.1.32:5000/%zz"} {
		r := Registry{Address: address, Username: "user", Password: "pass"}
		assert.Error(t, r.Validate(), address)
	}
}
