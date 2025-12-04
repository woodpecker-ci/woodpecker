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
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestServiceName(t *testing.T) {
	name, err := serviceName(&types.Step{Name: "database", UUID: "01he8bebctabr3kgk0qj36d2me"})
	assert.NoError(t, err)
	assert.Equal(t, "wp-svc-01he8bebctabr3kgk0qj36d2me-database", name)

	name, err = serviceName(&types.Step{Name: "wp-01he8bebctabr3kgk0qj36d2me-0-services-0.woodpecker-runtime.svc.cluster.local", UUID: "01he8bebctabr3kgk0qj36d2me"})
	assert.NoError(t, err)
	assert.Equal(t, "wp-svc-01he8bebctabr3kgk0qj36d2me-wp-01he8bebctabr3kgk0qj36d2me-0-services-0.woodpecker-runtime.svc.cluster.local", name)

	name, err = serviceName(&types.Step{Name: "awesome_service", UUID: "01he8bebctabr3kgk0qj36d2me"})
	assert.NoError(t, err)
	assert.Equal(t, "wp-svc-01he8bebctabr3kgk0qj36d2me-awesome-service", name)
}

func TestHeadlessService(t *testing.T) {
	expected := `
	{
	  "metadata": {
		"name": "wp-hsvc-11301",
		"namespace": "foo"
	  },
	  "spec": {
		"selector": {
		  "woodpecker-ci.org/task-uuid": "11301"
		},
		"clusterIP": "None",
		"type": "ClusterIP"
	  },
	  "status": {
		"loadBalancer": {}
	  }
	}`

	s, err := mkHeadlessService("foo", "11301")
	assert.NoError(t, err, "expected no error when creating headless service")
	j, err := json.Marshal(s)
	assert.NoError(t, err, "expected no error when marshaling headless service to JSON")
	assert.JSONEq(t, expected, string(j), "expected headless service JSON to match")
}

func TestInvalidHeadlessService(t *testing.T) {
	_, err := mkHeadlessService("foo", "invalid_task_uuid!")
	assert.Error(t, err, "expected error due to invalid task UUID")
}

func TestStartHeadlessService(t *testing.T) {
	t.Run("successfully creates headless service", func(t *testing.T) {
		engine := &kube{
			client: fake.NewClientset(),
			config: &config{Namespace: "test-namespace"},
		}

		svc, err := startHeadlessService(context.Background(), engine, "foo", "11301")
		assert.NoError(t, err, "expected no error when starting headless service")

		assert.NotNil(t, svc, "expected headless service to be created")
		assert.Equal(t, "wp-hsvc-11301", svc.Name, "expected headless service name to match")
		assert.Equal(t, "foo", svc.Namespace, "expected headless service namespace to match")
		assert.Equal(t, v1.ServiceTypeClusterIP, svc.Spec.Type, "expected headless service type to be ClusterIP")
		assert.Equal(t, "None", svc.Spec.ClusterIP, "expected headless service ClusterIP to be 'None'")
		assert.Equal(t, map[string]string{TaskUUIDLabel: "11301"}, svc.Spec.Selector)

		createdSvc, err := engine.client.CoreV1().Services("foo").Get(context.Background(), "wp-hsvc-11301", meta_v1.GetOptions{})
		assert.NoError(t, err, "expected no error when getting the created service")
		assert.Equal(t, svc.Name, createdSvc.Name, "expected created service name to match")
	})

	t.Run("error on invalid task UUID resulting in invalid domain-name", func(t *testing.T) {
		engine := &kube{
			client: fake.NewClientset(),
			config: &config{Namespace: "test-namespace"},
		}

		_, err := startHeadlessService(context.Background(), engine, "test-namespace", "invalid_task_uuid!")
		assert.Error(t, err, "expected error due to invalid task UUID")
	})
}

func TestStopHeadlessService(t *testing.T) {
	t.Run("successfully deletes headless service", func(t *testing.T) {
		engine := &kube{
			client: fake.NewClientset(),
			config: &config{Namespace: "test-namespace"},
		}

		// arrage
		_, err := startHeadlessService(context.Background(), engine, "foo", "11301")
		assert.NoError(t, err, "expected no error when starting headless service")

		_, err = engine.client.CoreV1().Services("foo").Get(context.Background(), "wp-hsvc-11301", meta_v1.GetOptions{})
		assert.NoError(t, err, "expected no error when getting the created service")

		// act
		err = stopHeadlessService(context.Background(), engine, "foo", "11301")
		assert.NoError(t, err, "expected no error when deleting headless service")

		// assert
		_, err = engine.client.CoreV1().Services("foo").Get(context.Background(), "wp-hsvc-11301", meta_v1.GetOptions{})
		assert.Error(t, err, "expected error when getting a deleted service")
		assert.True(t, err != nil, "expected error to be non-nil")
	})

	t.Run("handles non-existent service gracefully", func(t *testing.T) {
		engine := &kube{
			client: fake.NewClientset(),
			config: &config{Namespace: "test-namespace"},
		}

		err := stopHeadlessService(context.Background(), engine, "foo", "nonexistent")
		assert.NoError(t, err, "expected no error when deleting a non-existent service")
	})

	t.Run("error on invalid task UUID resulting in invalid domain-name", func(t *testing.T) {
		engine := &kube{
			client: fake.NewClientset(),
			config: &config{Namespace: "test-namespace"},
		}

		err := stopHeadlessService(context.Background(), engine, "test-namespace", "invalid_task_uuid!")
		assert.Error(t, err, "expected error due to invalid task UUID")
	})
}
