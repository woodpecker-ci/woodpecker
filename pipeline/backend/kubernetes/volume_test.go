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

	name, err = volumeName("woodpecker\\cache")
	assert.NoError(t, err)
	assert.Equal(t, "woodpecker\\cache", name)

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

	pvc, err := PersistentVolumeClaim("someNamespace", "somename", "local-storage", "1Gi", true)
	assert.NoError(t, err)

	j, err := json.Marshal(pvc)
	assert.NoError(t, err)
	assert.JSONEq(t, expectedRwx, string(j))

	pvc, err = PersistentVolumeClaim("someNamespace", "somename", "local-storage", "1Gi", false)
	assert.NoError(t, err)

	j, err = json.Marshal(pvc)
	assert.NoError(t, err)
	assert.JSONEq(t, expectedRwo, string(j))

	_, err = PersistentVolumeClaim("someNamespace", "some0INVALID3name", "local-storage", "1Gi", false)
	assert.Error(t, err)
}
