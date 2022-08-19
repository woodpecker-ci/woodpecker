package kubernetes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersistentVolumeClaim(t *testing.T) {
	expectedRwx := `
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

	expectedRwo := `
	{
	  "metadata": {
	    "name": "someName",
	    "namespace": "someNamespace",
	    "creationTimestamp": null
	  },
	  "spec": {
	    "accessModes": [
	      "ReadWriteOnce"
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

	pvc := PersistentVolumeClaim("someNamespace", "someName", "local-storage", "1Gi", true)
	j, err := json.Marshal(pvc)
	assert.Nil(t, err)
	assert.JSONEq(t, expectedRwx, string(j))

	pvc = PersistentVolumeClaim("someNamespace", "someName", "local-storage", "1Gi", false)
	j, err = json.Marshal(pvc)
	assert.Nil(t, err)
	assert.JSONEq(t, expectedRwo, string(j))
}
