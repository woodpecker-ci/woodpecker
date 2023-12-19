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
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

func TestServiceName(t *testing.T) {
	name, err := serviceName(&types.Step{Name: "wp_01he8bebctabr3kgk0qj36d2me_0_services_0"})
	assert.NoError(t, err)
	assert.Equal(t, "wp-01he8bebctabr3kgk0qj36d2me-0-services-0", name)

	name, err = serviceName(&types.Step{Name: "wp-01he8bebctabr3kgk0qj36d2me-0\\services-0"})
	assert.NoError(t, err)
	assert.Equal(t, "wp-01he8bebctabr3kgk0qj36d2me-0\\services-0", name)

	_, err = serviceName(&types.Step{Name: "wp-01he8bebctabr3kgk0qj36d2me-0-services-0.woodpecker-runtime.svc.cluster.local"})
	assert.ErrorIs(t, err, ErrDNSPatternInvalid)
}

func TestService(t *testing.T) {
	expected := `
	{
	  "metadata": {
	    "name": "bar",
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
	        "port": 2,
	        "targetPort": 2
	      },
	      {
	        "name": "port-3",
	        "port": 3,
	        "targetPort": 3
	      }
	    ],
	    "selector": {
	      "step": "baz"
	    },
	    "type": "ClusterIP"
	  },
	  "status": {
	    "loadBalancer": {}
	  }
	}`

	s, _ := Service("foo", "bar", []uint16{1, 2, 3}, map[string]string{"step": "baz"})
	j, err := json.Marshal(s)
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(j))
}
