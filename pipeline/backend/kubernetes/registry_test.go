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
	"encoding/json"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPullSecret(t *testing.T) {
	expected := `
	{
		"metadata": {
			"name": "5b4df028-55bd-4187-8291-4725e02fcea6",
			"namespace": "ns",
			"creationTimestamp": null
		},
		"data": {
			".dockerconfigjson": "eyJhdXRocyI6eyJnY3IuaW8iOnsiYXV0aCI6ImV5SjFjMlZ5Ym1GdFpTSTZJblZ6WlhJaUxDSndZWE56ZDI5eVpDSTZJbkJoYzNNaWZRPT0ifSwiaHViLmRvY2tlci5jb20iOnsiYXV0aCI6ImUzMD0ifX19"
		},
		"type": "kubernetes.io/dockerconfigjson"
	}`

	emptyRegistry := types.Registry{
		Hostname: "hub.docker.com",
	}
	normalRegistry := types.Registry{
		Hostname: "gcr.io",
		Username: "user",
		Password: "pass",
	}

	authsB64, err := dockerAuths([]*types.Registry{&emptyRegistry, &normalRegistry})
	assert.NoError(t, err)

	pullSecret, err := mkPullSecret("ns", "5b4df028-55bd-4187-8291-4725e02fcea6", authsB64)
	assert.NoError(t, err)

	pullSecretJson, err := json.Marshal(pullSecret)
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(pullSecretJson))
}
