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
	v1 "k8s.io/api/core/v1"

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
	name, err = stepToPodName(&types.Step{UUID: "01he8bebctabr3kg", Name: "cache", Type: types.StepTypeCache})
	assert.NoError(t, err)
	assert.EqualValues(t, "wp-01he8bebctabr3kg", name)
	name, err = stepToPodName(&types.Step{UUID: "01he8bebctabr3kg", Name: "release", Type: types.StepTypePlugin})
	assert.NoError(t, err)
	assert.EqualValues(t, "wp-01he8bebctabr3kg", name)
	name, err = stepToPodName(&types.Step{UUID: "01he8bebctabr3kg", Name: "prepare-env", Type: types.StepTypeCommands})
	assert.NoError(t, err)
	assert.EqualValues(t, "wp-01he8bebctabr3kg", name)
	name, err = stepToPodName(&types.Step{UUID: "01he8bebctabr3kg", Name: "postgres", Type: types.StepTypeService})
	assert.NoError(t, err)
	assert.EqualValues(t, "wp-svc-01he8bebctabr3kg-postgres", name)
}

func TestStepLabel(t *testing.T) {
	name, err := stepLabel(&types.Step{Name: "Build image"})
	assert.NoError(t, err)
	assert.EqualValues(t, "build-image", name)

	_, err = stepLabel(&types.Step{Name: ".build.image"})
	assert.ErrorIs(t, err, ErrDNSPatternInvalid)
}

func TestTinyPod(t *testing.T) {
	const expected = `
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
						"-c",
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
	}, "wp-01he8bebctabr3kgk0qj36d2me-0", "linux/amd64", BackendOptions{})
	assert.NoError(t, err)

	podJSON, err := json.Marshal(pod)
	assert.NoError(t, err)

	ja := jsonassert.New(t)
	ja.Assertf(string(podJSON), expected)
}

func TestFullPod(t *testing.T) {
	const expected = `
	{
		"metadata": {
			"name": "wp-01he8bebctabr3kgk0qj36d2me-0",
			"namespace": "woodpecker",
			"creationTimestamp": null,
			"labels": {
				"app": "test",
				"part-of": "woodpecker-ci",
				"step": "go-test"
			},
			"annotations": {
				"apps.kubernetes.io/pod-index": "0",
				"kubernetes.io/limit-ranger": "LimitRanger plugin set: cpu, memory request and limit for container"
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
				"storage": "ssd",
				"topology.kubernetes.io/region": "eu-central-1"
			},
			"runtimeClassName": "runc",
			"serviceAccountName": "wp-svc-acc",
			"securityContext": {
				"runAsUser": 101,
				"runAsGroup": 101,
				"runAsNonRoot": true,
				"fsGroup": 101,
				"appArmorProfile": {
					"type": "Localhost",
					"localhostProfile": "k8s-apparmor-example-deny-write"
				},
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
				},
				{
					"name": "wp-01he8bebctabr3kgk0qj36d2me-0"
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

	runtimeClass := "runc"
	hostAliases := []types.HostAlias{
		{Name: "cloudflare", IP: "1.1.1.1"},
		{Name: "cf.v6", IP: "2606:4700:4700::64"},
	}
	ports := []types.Port{
		{Number: 1234},
		{Number: 2345, Protocol: "tcp"},
		{Number: 3456, Protocol: "udp"},
	}
	secCtx := SecurityContext{
		Privileged:   newBool(true),
		RunAsNonRoot: newBool(true),
		RunAsUser:    newInt64(101),
		RunAsGroup:   newInt64(101),
		FSGroup:      newInt64(101),
		SeccompProfile: &SecProfile{
			Type:             "Localhost",
			LocalhostProfile: "profiles/audit.json",
		},
		ApparmorProfile: &SecProfile{
			Type:             "Localhost",
			LocalhostProfile: "k8s-apparmor-example-deny-write",
		},
	}
	pod, err := mkPod(&types.Step{
		UUID:        "01he8bebctabr3kgk0qj36d2me-0",
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
		AuthConfig: types.Auth{
			Username: "foo",
			Password: "bar",
		},
	}, &config{
		Namespace:                   "woodpecker",
		ImagePullSecretNames:        []string{"regcred", "another-pull-secret"},
		PodLabels:                   map[string]string{"app": "test"},
		PodLabelsAllowFromStep:      true,
		PodAnnotations:              map[string]string{"apps.kubernetes.io/pod-index": "0"},
		PodAnnotationsAllowFromStep: true,
		PodNodeSelector:             map[string]string{"topology.kubernetes.io/region": "eu-central-1"},
		SecurityContext:             SecurityContextConfig{RunAsNonRoot: false},
	}, "wp-01he8bebctabr3kgk0qj36d2me-0", "linux/amd64", BackendOptions{
		Labels:             map[string]string{"part-of": "woodpecker-ci"},
		Annotations:        map[string]string{"kubernetes.io/limit-ranger": "LimitRanger plugin set: cpu, memory request and limit for container"},
		NodeSelector:       map[string]string{"storage": "ssd"},
		RuntimeClassName:   &runtimeClass,
		ServiceAccountName: "wp-svc-acc",
		Tolerations:        []Toleration{{Key: "net-port", Value: "100Mbit", Effect: TaintEffectNoSchedule}},
		Resources: Resources{
			Requests: map[string]string{"memory": "128Mi", "cpu": "1000m"},
			Limits:   map[string]string{"memory": "256Mi", "cpu": "2"},
		},
		SecurityContext: &secCtx,
	})
	assert.NoError(t, err)

	podJSON, err := json.Marshal(pod)
	assert.NoError(t, err)

	ja := jsonassert.New(t)
	ja.Assertf(string(podJSON), expected)
}

func TestPodPrivilege(t *testing.T) {
	createTestPod := func(stepPrivileged, globalRunAsRoot bool, secCtx SecurityContext) (*v1.Pod, error) {
		return mkPod(&types.Step{
			Name:       "go-test",
			Image:      "golang:1.16",
			Privileged: stepPrivileged,
		}, &config{
			Namespace:       "woodpecker",
			SecurityContext: SecurityContextConfig{RunAsNonRoot: globalRunAsRoot},
		}, "wp-01he8bebctabr3kgk0qj36d2me-0", "linux/amd64", BackendOptions{
			SecurityContext: &secCtx,
		})
	}

	// securty context is requesting user and group 101 (non-root)
	secCtx := SecurityContext{
		RunAsUser:  newInt64(101),
		RunAsGroup: newInt64(101),
		FSGroup:    newInt64(101),
	}
	pod, err := createTestPod(false, false, secCtx)
	assert.NoError(t, err)
	assert.Equal(t, int64(101), *pod.Spec.SecurityContext.RunAsUser)
	assert.Equal(t, int64(101), *pod.Spec.SecurityContext.RunAsGroup)
	assert.Equal(t, int64(101), *pod.Spec.SecurityContext.FSGroup)

	// securty context is requesting root, but step is not privileged
	secCtx = SecurityContext{
		RunAsUser:  newInt64(0),
		RunAsGroup: newInt64(0),
		FSGroup:    newInt64(0),
	}
	pod, err = createTestPod(false, false, secCtx)
	assert.NoError(t, err)
	assert.Nil(t, pod.Spec.SecurityContext)
	assert.Nil(t, pod.Spec.Containers[0].SecurityContext)

	// step is not privileged, but security context is requesting privileged
	secCtx = SecurityContext{
		Privileged: newBool(true),
	}
	pod, err = createTestPod(false, false, secCtx)
	assert.NoError(t, err)
	assert.Nil(t, pod.Spec.SecurityContext)
	assert.Nil(t, pod.Spec.Containers[0].SecurityContext)

	// step is privileged and security context is requesting privileged
	secCtx = SecurityContext{
		Privileged: newBool(true),
	}
	pod, err = createTestPod(true, false, secCtx)
	assert.NoError(t, err)
	assert.True(t, *pod.Spec.Containers[0].SecurityContext.Privileged)

	// step is privileged and no security context is provided
	secCtx = SecurityContext{}
	pod, err = createTestPod(true, false, secCtx)
	assert.NoError(t, err)
	assert.True(t, *pod.Spec.Containers[0].SecurityContext.Privileged)

	// global runAsNonRoot is true and override is requested value by security context
	secCtx = SecurityContext{
		RunAsNonRoot: newBool(false),
	}
	pod, err = createTestPod(false, true, secCtx)
	assert.NoError(t, err)
	assert.True(t, *pod.Spec.SecurityContext.RunAsNonRoot)
}

func TestScratchPod(t *testing.T) {
	const expected = `
	{
		"metadata": {
			"name": "wp-01he8bebctabr3kgk0qj36d2me-0",
			"namespace": "woodpecker",
			"creationTimestamp": null,
			"labels": {
				"step": "curl-google"
			}
		},
		"spec": {
			"containers": [
				{
					"name": "wp-01he8bebctabr3kgk0qj36d2me-0",
					"image": "quay.io/curl/curl",
					"command": [
						"/usr/bin/curl",
						"-v",
						"google.com"
					],
					"resources": {}
				}
			],
			"restartPolicy": "Never"
		},
		"status": {}
	}`

	pod, err := mkPod(&types.Step{
		Name:       "curl-google",
		Image:      "quay.io/curl/curl",
		Entrypoint: []string{"/usr/bin/curl", "-v", "google.com"},
	}, &config{
		Namespace: "woodpecker",
	}, "wp-01he8bebctabr3kgk0qj36d2me-0", "linux/amd64", BackendOptions{})
	assert.NoError(t, err)

	podJSON, err := json.Marshal(pod)
	assert.NoError(t, err)

	ja := jsonassert.New(t)
	ja.Assertf(string(podJSON), expected)
}

func TestSecrets(t *testing.T) {
	const expected = `
	{
		"metadata": {
			"name": "wp-3kgk0qj36d2me01he8bebctabr-0",
			"namespace": "woodpecker",
			"creationTimestamp": null,
			"labels": {
				"step": "test-secrets"
			}
		},
		"spec": {
			"volumes": [
				{
					"name": "workspace",
					"persistentVolumeClaim": {
						"claimName": "workspace"
					}
				},
				{
					"name": "reg-cred",
					"secret": {
						"secretName": "reg-cred"
					}
				}
			],
			"containers": [
				{
					"name": "wp-3kgk0qj36d2me01he8bebctabr-0",
					"image": "alpine",
					"envFrom": [
						{
							"secretRef": {
								"name": "ghcr-push-secret"
							}
						}
					],
					"env": [
						{
							"name": "CGO",
							"value": "0"
						},
						{
							"name": "AWS_ACCESS_KEY_ID",
							"valueFrom": {
								"secretKeyRef": {
									"name": "aws-ecr",
									"key": "AWS_ACCESS_KEY_ID"
								}
							}
						},
						{
							"name": "AWS_SECRET_ACCESS_KEY",
							"valueFrom": {
								"secretKeyRef": {
									"name": "aws-ecr",
									"key": "access-key"
								}
							}
						}
					],
					"resources": {},
					"volumeMounts": [
						{
							"name": "workspace",
							"mountPath": "/woodpecker/src"
						},
						{
							"name": "reg-cred",
							"mountPath": "~/.docker/config.json",
							"subPath": ".dockerconfigjson",
							"readOnly": true
						}
					]
				}
			],
			"restartPolicy": "Never"
		},
		"status": {}
	}`

	pod, err := mkPod(&types.Step{
		Name:        "test-secrets",
		Image:       "alpine",
		Environment: map[string]string{"CGO": "0"},
		Volumes:     []string{"workspace:/woodpecker/src"},
	}, &config{
		Namespace:                  "woodpecker",
		NativeSecretsAllowFromStep: true,
	}, "wp-3kgk0qj36d2me01he8bebctabr-0", "linux/amd64", BackendOptions{
		Secrets: []SecretRef{
			{
				Name: "ghcr-push-secret",
			},
			{
				Name: "aws-ecr",
				Key:  "AWS_ACCESS_KEY_ID",
			},
			{
				Name:   "aws-ecr",
				Key:    "access-key",
				Target: SecretTarget{Env: "AWS_SECRET_ACCESS_KEY"},
			},
			{
				Name:   "reg-cred",
				Key:    ".dockerconfigjson",
				Target: SecretTarget{File: "~/.docker/config.json"},
			},
		},
	})
	assert.NoError(t, err)

	podJSON, err := json.Marshal(pod)
	assert.NoError(t, err)

	ja := jsonassert.New(t)
	ja.Assertf(string(podJSON), expected)
}
