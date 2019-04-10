package kubernetes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersistentVolume(t *testing.T) {
	expected := `
	{
	  "metadata": {
	    "name": "someName",
	    "namespace": "someNamespace",
	    "creationTimestamp": null
	  },
	  "spec": {
	    "capacity": {
	      "storage": "1Gi"
	    },
	    "local": {
	      "path": "/tmp"
	    },
	    "accessModes": [
	      "ReadWriteMany"
	    ],
	    "persistentVolumeReclaimPolicy": "Retain",
	    "storageClassName": "local-storage",
	    "nodeAffinity": {
	      "required": {
	        "nodeSelectorTerms": [
	          {
	            "matchExpressions": [
	              {
	                "key": "kubernetes.io/hostname",
	                "operator": "In",
	                "values": [
	                  "someNode"
	                ]
	              }
	            ]
	          }
	        ]
	      }
	    }
	  },
	  "status": {}
	}`

	pv := PersistentVolume("someNode", "someNamespace", "someName")
	j, err := json.Marshal(pv)
	assert.Nil(t, err)
	assert.JSONEq(t, expected, string(j))
}

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

	pvc := PersistentVolumeClaim("someNamespace", "someName")
	j, err := json.Marshal(pvc)
	assert.Nil(t, err)
	assert.JSONEq(t, expected, string(j))
}
