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
	"testing"

	"github.com/stretchr/testify/assert"

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
	    "namespace": "foo",
	    "creationTimestamp": null
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
