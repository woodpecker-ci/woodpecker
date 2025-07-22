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

func TestPvcName(t *testing.T) {
	name, err := volumeName("woodpecker_cache:/woodpecker/src/cache")
	assert.NoError(t, err)
	assert.Equal(t, "woodpecker-cache", name)

	_, err = volumeName("woodpecker\\cache")
	assert.ErrorIs(t, err, ErrDNSPatternInvalid)

	_, err = volumeName("-woodpecker.cache:/woodpecker/src/cache")
	assert.ErrorIs(t, err, ErrDNSPatternInvalid)
}

func TestPvcMount(t *testing.T) {
	mount := volumeMountPath("woodpecker-cache:/woodpecker/src/cache")
	assert.Equal(t, "/woodpecker/src/cache", mount)

	mount = volumeMountPath("/woodpecker/src/cache")
	assert.Equal(t, "/woodpecker/src/cache", mount)
}

func TestPersistentVolumeClaim(t *testing.T) {
	namespace := "someNamespace"
	expectedRwx := `
	{
	  "metadata": {
	    "name": "somename",
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
	    "name": "somename",
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

	pvc, err := mkPersistentVolumeClaim(&config{
		Namespace:    namespace,
		StorageClass: "local-storage",
		VolumeSize:   "1Gi",
		StorageRwx:   true,
	}, "somename", namespace)
	assert.NoError(t, err)

	j, err := json.Marshal(pvc)
	assert.NoError(t, err)
	assert.JSONEq(t, expectedRwx, string(j))

	pvc, err = mkPersistentVolumeClaim(&config{
		Namespace:    namespace,
		StorageClass: "local-storage",
		VolumeSize:   "1Gi",
		StorageRwx:   false,
	}, "somename", namespace)
	assert.NoError(t, err)

	j, err = json.Marshal(pvc)
	assert.NoError(t, err)
	assert.JSONEq(t, expectedRwo, string(j))

	_, err = mkPersistentVolumeClaim(&config{
		Namespace:    namespace,
		StorageClass: "local-storage",
		VolumeSize:   "1Gi",
		StorageRwx:   false,
	}, "some0..INVALID3name", namespace)
	assert.Error(t, err)
}
