// Copyright 2023 Woodpecker Authors
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
	"go.woodpecker-ci.org/woodpecker/pipeline/backend/types"
)

func TestPodName(t *testing.T) {
	name, err := podName(&types.Step{Name: "wp_01he8bebctabr3kgk0qj36d2me_0"})
	assert.NoError(t, err)
	assert.Equal(t, "wp-01he8bebctabr3kgk0qj36d2me-0", name)

	name, err = podName(&types.Step{Name: "wp\\01he8bebctabr3kgk0qj36d2me-0"})
	assert.NoError(t, err)
	assert.Equal(t, "wp\\01he8bebctabr3kgk0qj36d2me-0", name)

	_, err = podName(&types.Step{Name: "wp-01he8bebctabr3kgk0qj36d2me-0-services-0.woodpecker-runtime.svc.cluster.local"})
	assert.ErrorIs(t, err, ErrDNSPatternInvalid)
}

func TestTinyPod(t *testing.T) {
	expected := `
	{
		"metadata": {
			"name": "wp-01he8bebctabr3kgk0qj36d2me-0",
			"namespace": "woodpecker",
			"creationTimestamp": null,
			"labels": {
				"step": "wp-01he8bebctabr3kgk0qj36d2me-0"
			}
		},
		"spec": {
			"volumes": [
				{
					"name": "workspace",
					"persistentVolumeClaim": {
						"claimName": "workspace"
					}
				}
			],
			"containers": [
				{
					"name": "wp-01he8bebctabr3kgk0qj36d2me-0",
					"image": "gradle:8.4.0-jdk21",
					"command": [
						"/bin/sh",
						"-c"
					],
					"args": [
						"echo $CI_SCRIPT | base64 -d | /bin/sh -e"
					],
					"workingDir": "/woodpecker/src",
					"env": [
						{
							"name": "CI",
							"value": "woodpecker"
						},
						{
							"name": "HOME",
							"value": "/root"
						},
						{
							"name": "SHELL",
							"value": "/bin/sh"
						},
						{
							"name": "CI_SCRIPT",
							"value": "CmlmIFsgLW4gIiRDSV9ORVRSQ19NQUNISU5FIiBdOyB0aGVuCmNhdCA8PEVPRiA+ICRIT01FLy5uZXRyYwptYWNoaW5lICRDSV9ORVRSQ19NQUNISU5FCmxvZ2luICRDSV9ORVRSQ19VU0VSTkFNRQpwYXNzd29yZCAkQ0lfTkVUUkNfUEFTU1dPUkQKRU9GCmNobW9kIDA2MDAgJEhPTUUvLm5ldHJjCmZpCnVuc2V0IENJX05FVFJDX1VTRVJOQU1FCnVuc2V0IENJX05FVFJDX1BBU1NXT1JECnVuc2V0IENJX1NDUklQVAoKZWNobyArICdncmFkbGUgYnVpbGQnCmdyYWRsZSBidWlsZAo="
						}
					],
					"resources": {},
					"volumeMounts": [
						{
							"name": "workspace",
							"mountPath": "/woodpecker/src"
						}
					],
					"securityContext": {
						"privileged": false
					}
				}
			],
			"restartPolicy": "Never",
			"imagePullSecrets": [
				{
					"name": "regcred"
				}
			]
		},
		"status": {}
	}`

	pod, err := Pod("woodpecker", "wp-01he8bebctabr3kgk0qj36d2me-0", "gradle:8.4.0-jdk21", "/woodpecker/src", "linux/amd64", "",
		false, false,
		[]string{"gradle build"}, []string{"workspace:/woodpecker/src"}, nil,
		nil, nil, map[string]string{"CI": "woodpecker"}, nil,
		nil,
		types.Resources{Requests: nil, Limits: nil},
	)
	assert.NoError(t, err)
	json, err := json.Marshal(pod)
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(json))
}

func TestFullPod(t *testing.T) {
	expected := `
	{
		"metadata": {
			"name": "wp-01he8bebctabr3kgk0qj36d2me-0",
			"namespace": "woodpecker",
			"creationTimestamp": null,
			"labels": {
				"app": "test",
				"step": "wp-01he8bebctabr3kgk0qj36d2me-0"
			},
			"annotations": {
				"apparmor.security": "runtime/default"
			}
		},
		"spec": {
			"volumes": [
				{
					"name": "woodpecker-cache",
					"persistentVolumeClaim": {
						"claimName": "woodpecker-cache"
					}
				}
			],
			"containers": [
				{
					"name": "wp-01he8bebctabr3kgk0qj36d2me-0",
					"image": "meltwater/drone-cache",
					"command": [
						"/bin/sh",
						"-c"
					],
					"args": [
						"echo $CI_SCRIPT | base64 -d | /bin/sh -e"
					],
					"workingDir": "/woodpecker/src",
					"env": [
						{
							"name": "CGO",
							"value": "0"
						},
						{
							"name": "CI_SCRIPT",
							"value": "CmlmIFsgLW4gIiRDSV9ORVRSQ19NQUNISU5FIiBdOyB0aGVuCmNhdCA8PEVPRiA+ICRIT01FLy5uZXRyYwptYWNoaW5lICRDSV9ORVRSQ19NQUNISU5FCmxvZ2luICRDSV9ORVRSQ19VU0VSTkFNRQpwYXNzd29yZCAkQ0lfTkVUUkNfUEFTU1dPUkQKRU9GCmNobW9kIDA2MDAgJEhPTUUvLm5ldHJjCmZpCnVuc2V0IENJX05FVFJDX1VTRVJOQU1FCnVuc2V0IENJX05FVFJDX1BBU1NXT1JECnVuc2V0IENJX1NDUklQVAoKZWNobyArICdnbyBnZXQnCmdvIGdldAoKZWNobyArICdnbyB0ZXN0JwpnbyB0ZXN0Cg=="
						},
						{
							"name": "HOME",
							"value": "/root"
						},
						{
							"name": "SHELL",
							"value": "/bin/sh"
						}
					],
					"resources": {
						"limits": {
							"cpu": "2",
							"memory": "256Mi"
						},
						"requests": {
							"cpu": "1",
							"memory": "128Mi"
						}
					},
					"volumeMounts": [
						{
							"name": "woodpecker-cache",
							"mountPath": "/woodpecker/src/cache"
						}
					],
					"imagePullPolicy": "Always",
					"securityContext": {
						"privileged": true
					}
				}
			],
			"restartPolicy": "Never",
			"nodeSelector": {
				"storage": "ssd"
			},
			"serviceAccountName": "wp-svc-acc",
			"imagePullSecrets": [
				{
					"name": "regcred"
				}
			],
			"tolerations": [
				{
					"key": "net-port",
					"value": "100Mbit",
					"effect": "NoSchedule"
				}
			],
			"hostAliases": [
				{
					"ip": "1.1.1.1",
					"hostnames": [
						"cloudflare"
					]
				}
			]
		},
		"status": {}
	}`

	pod, err := Pod("woodpecker", "wp-01he8bebctabr3kgk0qj36d2me-0", "meltwater/drone-cache", "/woodpecker/src", "linux/amd64", "wp-svc-acc",
		true, true,
		[]string{"go get", "go test"}, []string{"woodpecker-cache:/woodpecker/src/cache"}, []string{"cloudflare:1.1.1.1"},
		map[string]string{"app": "test"}, map[string]string{"apparmor.security": "runtime/default"}, map[string]string{"CGO": "0"}, map[string]string{"storage": "ssd"},
		[]types.Toleration{{Key: "net-port", Value: "100Mbit", Effect: types.TaintEffectNoSchedule}},
		types.Resources{Requests: map[string]string{"memory": "128Mi", "cpu": "1000m"}, Limits: map[string]string{"memory": "256Mi", "cpu": "2"}},
	)
	assert.NoError(t, err)
	json, err := json.Marshal(pod)
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(json))
}
