package kubernetes

import (
	"encoding/json"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

func TestNoAuthNoSecret(t *testing.T) {
	assert.False(t, needsImagePullSecret(&types.Step{}))
}

func TestNoPasswordNoSecret(t *testing.T) {
	assert.False(t, needsImagePullSecret(&types.Step{
		AuthConfig: types.Auth{Username: "foo"},
	}))
}

func TestNoUsernameNoSecret(t *testing.T) {
	assert.False(t, needsImagePullSecret(&types.Step{
		AuthConfig: types.Auth{Password: "foo"},
	}))
}

func TestUsernameAndPasswordNeedsSecret(t *testing.T) {
	assert.True(t, needsImagePullSecret(&types.Step{
		AuthConfig: types.Auth{Username: "foo", Password: "bar"},
	}))
}

func TestImagePullSecret(t *testing.T) {
	expected := `{
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
  	"type": "kubernetes.io/dockerconfigjson",
	  "data": {
			".dockerconfigjson": "eyJhdXRocyI6eyJkb2NrZXIuaW8iOnsidXNlcm5hbWUiOiJmb28iLCJwYXNzd29yZCI6ImJhciJ9fX0="
	  }
	}`

	pullSecret, err := mkImagePullSecret(&types.Step{
		Name:        "go-test",
		Image:       "meltwater/drone-cache",
		WorkingDir:  "/woodpecker/src",
		Pull:        true,
		Privileged:  true,
		Commands:    []string{"go get", "go test"},
		Entrypoint:  []string{"/bin/sh", "-c"},
		Volumes:     []string{"woodpecker-cache:/woodpecker/src/cache"},
		Environment: map[string]string{"CGO": "0"},
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
		ServiceAccountName: "wp-svc-acc",
		Tolerations:        []Toleration{{Key: "net-port", Value: "100Mbit", Effect: TaintEffectNoSchedule}},
		Resources: Resources{
			Requests: map[string]string{"memory": "128Mi", "cpu": "1000m"},
			Limits:   map[string]string{"memory": "256Mi", "cpu": "2"},
		},
	})
	assert.NoError(t, err)

	podJSON, err := json.Marshal(pullSecret)
	assert.NoError(t, err)

	ja := jsonassert.New(t)
	ja.Assertf(string(podJSON), expected)
}
