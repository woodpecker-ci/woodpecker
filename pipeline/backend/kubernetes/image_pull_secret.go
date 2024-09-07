package kubernetes

import (
	"encoding/json"

	"github.com/distribution/reference"
	config_file "github.com/docker/cli/cli/config/configfile"
	config_file_types "github.com/docker/cli/cli/config/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/utils"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func needsImagePullSecret(step *types.Step) bool {
	return step.AuthConfig.Username != "" && step.AuthConfig.Password != ""
}

func mkImagePullSecret(step *types.Step, config *config, podName, goos string, options BackendOptions) (*v1.Secret, error) {
	labels, err := podLabels(step, config, options)
	if err != nil {
		return nil, err
	}
	annotations := podAnnotations(config, options)

	named, err := utils.ParseNamed(step.Image)
	if err != nil {
		return nil, err
	}

	authConfig := config_file.ConfigFile{
		AuthConfigs: map[string]config_file_types.AuthConfig{
			reference.Domain(named): {
				Username: step.AuthConfig.Username,
				Password: step.AuthConfig.Password,
			},
		},
	}

	configFileJSON, err := json.Marshal(authConfig)
	if err != nil {
		return nil, err
	}

	return &v1.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Namespace:   config.Namespace,
			Name:        podName,
			Labels:      labels,
			Annotations: annotations,
		},
		Type: v1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			v1.DockerConfigJsonKey: configFileJSON,
		},
	}, nil
}
