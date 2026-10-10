// Copyright 2026 Woodpecker Authors
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

package model

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentSnapshot(t *testing.T) {
	agent := &Agent{
		ID:           3,
		Name:         "builder-03",
		OwnerID:      7,
		OrgID:        5,
		Token:        "secret-token",
		Platform:     "linux/amd64",
		Backend:      "docker",
		Version:      "3.19.0",
		CustomLabels: map[string]string{"zone": "eu"},
		Filters:      map[string]string{"repo": "a/b"},
	}

	snapshot := NewAgentSnapshot(agent)

	// the agent id is not part of the snapshot, the workflow holds it
	assert.Equal(t, &AgentSnapshot{
		OrgID:        5,
		Name:         "builder-03",
		Platform:     "linux/amd64",
		Backend:      "docker",
		CustomLabels: map[string]string{"zone": "eu"},
	}, snapshot)

	// later changes of the agent must not change the snapshot
	agent.CustomLabels["zone"] = "us"
	assert.Equal(t, "eu", snapshot.CustomLabels["zone"])

	raw, err := json.Marshal(snapshot)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "secret-token")
}

func TestAgentSnapshotContentHash(t *testing.T) {
	base := func() *AgentSnapshot {
		return &AgentSnapshot{
			OrgID:        5,
			Name:         "builder",
			Platform:     "linux/amd64",
			Backend:      "docker",
			CustomLabels: map[string]string{"zone": "eu"},
		}
	}
	baseHash, err := base().ContentHash()
	require.NoError(t, err)
	assert.Len(t, baseHash, 64)

	// walk all fields, so a field added later cannot be forgotten: every stored field must
	// change the hash, the rest must not
	typ := reflect.TypeFor[AgentSnapshot]()
	for i := range typ.NumField() {
		field := typ.Field(i)
		t.Run(field.Name, func(t *testing.T) {
			snapshot := base()
			value := reflect.ValueOf(snapshot).Elem().Field(i)
			switch value.Kind() {
			case reflect.Int64:
				value.SetInt(value.Int() + 42)
			case reflect.String:
				value.SetString(value.String() + "-changed")
			case reflect.Map:
				value.Set(reflect.ValueOf(map[string]string{"zone": "us"}))
			default:
				t.Fatalf("field %s of kind %s is not covered by this test", field.Name, value.Kind())
			}

			hash, err := snapshot.ContentHash()
			require.NoError(t, err)

			stored := field.Tag.Get("xorm") != "-"
			switch field.Name {
			case "ID", "Hash":
				// stored, but they are the row's identity, not its content
				stored = false
			}
			if stored {
				assert.NotEqual(t, baseHash, hash, "stored field %s must be part of the hash", field.Name)
			} else {
				assert.Equal(t, baseHash, hash, "field %s must not be part of the hash", field.Name)
			}
		})
	}

	t.Run("missing and empty labels are the same", func(t *testing.T) {
		withNil, withEmpty := base(), base()
		withNil.CustomLabels = nil
		withEmpty.CustomLabels = map[string]string{}
		nilHash, err := withNil.ContentHash()
		require.NoError(t, err)
		emptyHash, err := withEmpty.ContentHash()
		require.NoError(t, err)
		assert.Equal(t, nilHash, emptyHash)
	})
}
