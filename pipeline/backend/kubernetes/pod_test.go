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

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

func TestPodName(t *testing.T) {
	name, err := podName(&types.Step{UUID: "01he8bebctabr3kgk0qj36d2me-0"})
	assert.NoError(t, err)
	assert.Equal(t, "wp-01he8bebctabr3kgk0qj36d2me-0", name)

	_, err = podName(&types.Step{UUID: "01he8bebctabr3kgk0qj36d2me\\0a"})
	assert.ErrorIs(t, err, ErrDNSPatternInvalid)

	_, err = podName(&types.Step{UUID: "01he8bebctabr3kgk0qj36d2me-0-services-0..woodpecker-runtime.svc.cluster.local"})
	assert.ErrorIs(t, err, ErrDNSPatternInvalid)
}

func TestStepToPodName(t *testing.T) {
	name, err := stepToPodName(&types.Step{UUID: "01he8bebctabr3kg", Name: "clone", Type: types.StepTypeClone})
	assert.NoError(t, err)
	assert.EqualValues(t, "wp-01he8bebctabr3kg", name)
	name, err = stepToPodName(&types.Step{UUID: "01he8bebctabr3kg", Name: "clone", Type: types.StepTypeCache})
	assert.NoError(t, err)
	assert.EqualValues(t, "wp-01he8bebctabr3kg", name)
	name, err = stepToPodName(&types.Step{UUID: "01he8bebctabr3kg", Name: "clone", Type: types.StepTypePlugin})
	assert.NoError(t, err)
	assert.EqualValues(t, "wp-01he8bebctabr3kg", name)
	name, err = stepToPodName(&types.Step{UUID: "01he8bebctabr3kg", Name: "clone", Type: types.StepTypeCommands})
	assert.NoError(t, err)
	assert.EqualValues(t, "wp-01he8bebctabr3kg", name)
	name, err = stepToPodName(&types.Step{UUID: "01he8bebctabr3kg", Name: "clone", Type: types.StepTypeService})
	assert.NoError(t, err)
	assert.EqualValues(t, "clone", name)
}

func TestTinyPod(t *testing.T) {
	expected := `
	{
		"metadata": {
			"name": "wp-01he8bebctabr3kgk0qj36d2me-0",
			"namespace": "woodpecker",
			"creationTimestamp": null,
			"labels": {
				"step": "build-via-gradle"
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
						"<<UNORDERED>>",
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
					]
				}
			],
			"restartPolicy": "Never"
		},
		"status": {}
	}`

	pod, err := mkPod(&types.Step{
		Name:        "build-via-gradle",
		Image:       "gradle:8.4.0-jdk21",
		WorkingDir:  "/woodpecker/src",
		Pull:        false,
		Privileged:  false,
		Commands:    []string{"gradle build"},
		Volumes:     []string{"workspace:/woodpecker/src"},
		Environment: map[string]string{"CI": "woodpecker"},
	}, &config{
		Namespace: "woodpecker",
	}, "wp-01he8bebctabr3kgk0qj36d2me-0", "linux/amd64")
	assert.NoError(t, err)

	json, err := json.Marshal(pod)
	assert.NoError(t, err)

	ja := jsonassert.New(t)
	ja.Assertf(string(json), expected)
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
				"step": "go-test"
			},
			"annotations": {
				"apps.kubernetes.io/pod-index": "0",
				"container.apparmor.security.beta.kubernetes.io/wp-01he8bebctabr3kgk0qj36d2me-0": "localhost/k8s-apparmor-example-deny-write"
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
					"ports": [
						{
							"containerPort": 1234
						},
						{
							"containerPort": 2345,
							"protocol": "TCP"
						},
						{
							"containerPort": 3456,
							"protocol": "UDP"
						}
					],
					"env": [
						"<<UNORDERED>>",
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
			"securityContext": {
				"runAsUser": 101,
				"runAsGroup": 101,
				"runAsNonRoot": true,
				"fsGroup": 101,
				"seccompProfile": {
        	"type": "Localhost",
          "localhostProfile": "profiles/audit.json"
				}
			},
			"imagePullSecrets": [
				{
					"name": "regcred"
				},
				{
					"name": "another-pull-secret"
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
				},
				{
					"ip": "2606:4700:4700::64",
					"hostnames": [
						"cf.v6"
					]
				}
			]
		},
		"status": {}
	}`

	hostAliases := []types.HostAlias{
		{Name: "cloudflare", IP: "1.1.1.1"},
		{Name: "cf.v6", IP: "2606:4700:4700::64"},
	}
	ports := []types.Port{
		{Number: 1234},
		{Number: 2345, Protocol: "tcp"},
		{Number: 3456, Protocol: "udp"},
	}
	secCtx := types.SecurityContext{
		Privileged:   newBool(true),
		RunAsNonRoot: newBool(true),
		RunAsUser:    newInt64(101),
		RunAsGroup:   newInt64(101),
		FSGroup:      newInt64(101),
		SeccompProfile: &types.SecProfile{
			Type:             "Localhost",
			LocalhostProfile: "profiles/audit.json",
		},
		ApparmorProfile: &types.SecProfile{
			Type:             "Localhost",
			LocalhostProfile: "k8s-apparmor-example-deny-write",
		},
	}
	pod, err := mkPod(&types.Step{
		Name:        "go-test",
		Image:       "meltwater/drone-cache",
		WorkingDir:  "/woodpecker/src",
		Pull:        true,
		Privileged:  true,
		Commands:    []string{"go get", "go test"},
		Entrypoint:  []string{"/bin/sh", "-c"},
		Volumes:     []string{"woodpecker-cache:/woodpecker/src/cache"},
		Environment: map[string]string{"CGO": "0"},
		ExtraHosts:  hostAliases,
		Ports:       ports,
		BackendOptions: types.BackendOptions{
			Kubernetes: types.KubernetesBackendOptions{
				NodeSelector:       map[string]string{"storage": "ssd"},
				ServiceAccountName: "wp-svc-acc",
				Tolerations:        []types.Toleration{{Key: "net-port", Value: "100Mbit", Effect: types.TaintEffectNoSchedule}},
				Resources: types.Resources{
					Requests: map[string]string{"memory": "128Mi", "cpu": "1000m"},
					Limits:   map[string]string{"memory": "256Mi", "cpu": "2"},
				},
				SecurityContext: &secCtx,
			},
		},
	}, &config{
		Namespace:            "woodpecker",
		ImagePullSecretNames: []string{"regcred", "another-pull-secret"},
		PodLabels:            map[string]string{"app": "test"},
		PodAnnotations:       map[string]string{"apps.kubernetes.io/pod-index": "0"},
		SecurityContext:      SecurityContextConfig{RunAsNonRoot: false},
	}, "wp-01he8bebctabr3kgk0qj36d2me-0", "linux/amd64")
	assert.NoError(t, err)

	json, err := json.Marshal(pod)
	assert.NoError(t, err)

	ja := jsonassert.New(t)
	ja.Assertf(string(json), expected)
}
