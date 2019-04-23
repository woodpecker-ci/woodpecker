package kubernetes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersistentVolumeClaim(t *testing.T) {
	expected := `
	{
	  "metadata": {
	    "name": "someName",
	    "namespace": "someNamespace",
	    "creationTimestamp": null
	  },
	  "spec": {
	    "accessModes": [
	      "ReadWriteMany"
	    ],
	    "resources": {
	      "requests": {
	        "storage": "1Gi"
	      }
	    },
	    "storageClassName": "local-storage"
	  },
	  "status": {}
	}`

	pvc := PersistentVolumeClaim("someNamespace", "someName", "local-storage", "1Gi")
	j, err := json.Marshal(pvc)
	assert.Nil(t, err)
	assert.JSONEq(t, expected, string(j))
}
