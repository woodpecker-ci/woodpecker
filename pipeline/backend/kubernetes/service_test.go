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
	assert.Nil(t, err)
	assert.JSONEq(t, expected, string(j))
}
