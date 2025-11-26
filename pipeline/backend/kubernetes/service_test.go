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

func TestService(t *testing.T) {
	expected := `
	{
	  "metadata": {
	    "name": "wp-svc-01he8bebctabr3kgk0qj36d2me-0-bar",
	    "namespace": "foo"
	  },
	  "spec": {
	    "ports": [
	      {
	        "name": "port-1",
	        "port": 1,
	        "targetPort": 1
	      },
	      {
	        "name": "port-2",
	        "protocol": "TCP",
	        "port": 2,
	        "targetPort": 2
	      },
	      {
	        "name": "port-3",
	        "protocol": "UDP",
	        "port": 3,
	        "targetPort": 3
	      }
	    ],
	    "selector": {
	      "service": "wp-svc-01he8bebctabr3kgk0qj36d2me-0-bar"
	    },
	    "type": "ClusterIP"
	  },
	  "status": {
	    "loadBalancer": {}
	  }
	}`
	ports := []types.Port{
		{Number: 1},
		{Number: 2, Protocol: "tcp"},
		{Number: 3, Protocol: "udp"},
	}
	s, err := mkService(&types.Step{
		Name:  "bar",
		UUID:  "01he8bebctabr3kgk0qj36d2me-0",
		Ports: ports,
	}, &config{Namespace: "foo"})
	assert.NoError(t, err)
	j, err := json.Marshal(s)
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(j))
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
	assert.NoError(t, err)
	j, err := json.Marshal(s)
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(j))
}

func TestInvalidHeadlessService(t *testing.T) {
	_, err := mkHeadlessService("foo", "invalid_task_uuid!")
	assert.Error(t, err)
}

func TestStartHeadlessService(t *testing.T) {
	t.Run("successfully creates headless service", func(t *testing.T) {
		engine := &kube{
			client: fake.NewClientset(),
			config: &config{Namespace: "test-namespace"},
		}

		svc, err := startHeadlessService(context.Background(), engine, "foo", "11301")
		assert.NoError(t, err)

		assert.NotNil(t, svc)
		assert.Equal(t, "wp-hsvc-11301", svc.Name)
		assert.Equal(t, "foo", svc.Namespace)
		assert.Equal(t, v1.ServiceTypeClusterIP, svc.Spec.Type)
		assert.Equal(t, "None", svc.Spec.ClusterIP)
		assert.Equal(t, map[string]string{TaskUUIDLabel: "11301"}, svc.Spec.Selector)

		createdSvc, err := engine.client.CoreV1().Services("foo").Get(context.Background(), "wp-hsvc-11301", meta_v1.GetOptions{})
		assert.NoError(t, err)
		assert.Equal(t, svc.Name, createdSvc.Name)
	})

	t.Run("error on invalid task UUID resulting in invalid domain-name", func(t *testing.T) {
		engine := &kube{
			client: fake.NewClientset(),
			config: &config{Namespace: "test-namespace"},
		}

		_, err := startHeadlessService(context.Background(), engine, "test-namespace", "invalid_task_uuid!")
		assert.Error(t, err)
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
		assert.NoError(t, err)

		// act
		err = stopHeadlessService(context.Background(), engine, "foo", "11301")
		assert.NoError(t, err)

		// assert
		_, err = engine.client.CoreV1().Services("foo").Get(context.Background(), "wp-hsvc-11301", meta_v1.GetOptions{})
		assert.Error(t, err)
		assert.True(t, err != nil)
	})

	t.Run("handles non-existent service gracefully", func(t *testing.T) {
		engine := &kube{
			client: fake.NewClientset(),
			config: &config{Namespace: "test-namespace"},
		}

		err := stopHeadlessService(context.Background(), engine, "foo", "nonexistent")
		assert.NoError(t, err)
	})

	t.Run("error on invalid task UUID resulting in invalid domain-name", func(t *testing.T) {
		engine := &kube{
			client: fake.NewClientset(),
			config: &config{Namespace: "test-namespace"},
		}

		err := stopHeadlessService(context.Background(), engine, "test-namespace", "invalid_task_uuid!")
		assert.Error(t, err)
	})
}
