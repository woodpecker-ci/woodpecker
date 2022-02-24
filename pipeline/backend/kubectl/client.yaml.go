package kubectl

import (
	"context"
	"errors"
	"io/ioutil"
	"os"

	"github.com/rs/zerolog/log"
)

// Deploy kubernetes yaml (apply, create, delete)
func (client *KubeClient) DeployKubectlYaml(
	command,
	yaml string,
	wait bool,
) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), client.RequestTimeout)
	out, err := client.DeployKubectlYamlWithContext(ctx, command, yaml, wait)
	cancel()
	return out, err
}

// Deploy kubernetes yaml in a specific context (apply, create, delete)
func (client *KubeClient) DeployKubectlYamlWithContext(
	ctx context.Context,
	command,
	yaml string,
	wait bool,
) (string, error) {
	yamlFile, err := ioutil.TempFile(os.TempDir(), "wp.setup.kubectl.*.yaml")
	if err != nil {
		return "", err
	}
	yamlFilename := yamlFile.Name()

	defer func() {
		err := os.Remove(yamlFilename)
		if err != nil {
			log.Error().
				Str("Path", yamlFilename).
				Err(err).
				Msg("Failed to remove yaml temp. File still exists.")
		}
	}()

	_, err = yamlFile.WriteString(yaml)
	if err != nil {
		return "", err
	}
	err = yamlFile.Close()
	if err != nil {
		return "", err
	}

	output, err := client.RunKubectlCommand(
		ctx,
		command,
		Triary(command == "delete", "--ignore-not-found=true", ""),
		Triary(wait, "--wait=true", "--wait=false"),
		"-f", yamlFilename,
	)

	if err != nil {
		err = errors.New("Failed to deploy yaml. " + err.Error())
	}

	return output, err
}
