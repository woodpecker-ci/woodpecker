// Copyright 2024 Woodpecker Authors
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDNSName(t *testing.T) {
	name, err := dnsName("wp_01he8bebctabr3kgk0qj36d2me_0_services_0")
	assert.NoError(t, err)
	assert.Equal(t, "wp-01he8bebctabr3kgk0qj36d2me-0-services-0", name)

	name, err = dnsName("a.0-AA")
	assert.NoError(t, err)
	assert.Equal(t, "a.0-aa", name)

	name, err = dnsName("wp-01he8bebctabr3kgk0qj36d2me-0-services-0.woodpecker-runtime.svc.cluster.local")
	assert.NoError(t, err)
	assert.Equal(t, "wp-01he8bebctabr3kgk0qj36d2me-0-services-0.woodpecker-runtime.svc.cluster.local", name)

	_, err = dnsName(".0-a")
	assert.ErrorContains(t, err, "name is not a valid kubernetes DNS name")

	_, err = dnsName("ABC..DEF")
	assert.ErrorContains(t, err, "name is not a valid kubernetes DNS name")

	_, err = dnsName("0.-a")
	assert.ErrorContains(t, err, "name is not a valid kubernetes DNS name")

	_, err = dnsName("test-")
	assert.ErrorContains(t, err, "name is not a valid kubernetes DNS name")

	_, err = dnsName("-test")
	assert.ErrorContains(t, err, "name is not a valid kubernetes DNS name")

	_, err = dnsName("0-a.")
	assert.ErrorContains(t, err, "name is not a valid kubernetes DNS name")

	_, err = dnsName("abc\\def")
	assert.ErrorContains(t, err, "name is not a valid kubernetes DNS name")
}

func TestToDnsName(t *testing.T) {
	name, err := toDNSName("BUILD_AND_DEPLOY_0")
	assert.NoError(t, err)
	assert.Equal(t, "build-and-deploy-0", name)

	name, err = toDNSName("build and deploy")
	assert.NoError(t, err)
	assert.Equal(t, "build-and-deploy", name)

	name, err = toDNSName("build & deploy")
	assert.NoError(t, err)
	assert.Equal(t, "build--deploy", name)

	_, err = toDNSName("-build-and-deploy")
	assert.ErrorContains(t, err, "name is not a valid kubernetes DNS name")
}
