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
)

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
	        "port": 1,
	        "targetPort": 1
	      },
	      {
	        "port": 2,
	        "targetPort": 2
	      },
	      {
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

	s, _ := Service("foo", "bar", "baz", []string{"1", "2", "3"})
	j, err := json.Marshal(s)
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(j))
}
