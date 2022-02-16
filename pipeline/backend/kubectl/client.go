package kubectl

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
)

type KubeCtlClientCoreArgs struct {
	Namespace string // the default namespace
	Context   string // the default context
}

func (this *KubeCtlClientCoreArgs) ToArgsList() []string {
	cmnd := []string{}
	if len(this.Namespace) > 0 {
		cmnd = append(cmnd, "--namespace", this.Namespace)
	}
	if len(this.Context) > 0 {
		cmnd = append(cmnd, "--context", this.Context)
	}
	return cmnd
}

func firstNotNil(args ...interface{}) interface{} {
	for _, arg := range args {
		switch arg.(type) {
		case string:
			if len(args) > 0 {
				return arg
			}
			break
		default:
			return arg
		}
	}
	return nil
}

func (this *KubeCtlClientCoreArgs) Merge(args KubeCtlClientCoreArgs) KubeCtlClientCoreArgs {
	return KubeCtlClientCoreArgs{
		Namespace: firstNotNil(args.Namespace, this.Namespace).(string),
		Context:   firstNotNil(args.Context, this.Context).(string),
	}
}

type KubeCtlClient struct {
	Executable string                // the default executable
	CoreArgs   KubeCtlClientCoreArgs // the default args
}

func (this *KubeCtlClient) GetExecutable() string {
	if len(this.Executable) == 0 {
		return "kubectl"
	}
	return this.Executable
}

// run a kubectl command
func (e *KubeCtlClient) RunKubectlCommand(args ...interface{}) (string, error) {
	rslt, err := exec.Command(e.GetExecutable(), e.ComposeKubectlCommand(args...)...).CombinedOutput()
	if err != nil {
		return "", errors.New(string(rslt) + err.Error())
	}
	return string(rslt), err
}

// create a kubectl command context
func (e *KubeCtlClient) GetKubectlCommandContext(ctx context.Context, args ...interface{}) *exec.Cmd {
	return exec.CommandContext(ctx, e.GetExecutable(), e.ComposeKubectlCommand(args...)...)
}

func (this *KubeCtlClient) ComposeKubectlCommand(args ...interface{}) []string {
	command := []string{}
	for _, ar := range args {
		switch ar.(type) {
		case string:
			if len(ar.(string)) == 0 {
				continue
			}
			command = append(command, ar.(string))
			break
		case []string:
			for _, part := range ar.([]string) {
				if len(part) == 0 {
					continue
				}
				command = append(command, part)
			}
			break
		default:
			continue
		}
	}

	return command
}

func (this *KubeCtlClient) DeployKubectlYaml(command, yaml string) (string, error) {
	yamlFile, err := ioutil.TempFile(os.TempDir(), "wp.setup.kubectl.*.bat")
	if err != nil {
		return "", err
	}

	_, err = yamlFile.WriteString(yaml)
	if err != nil {
		return "", err
	}
	err = yamlFile.Close()
	if err != nil {
		return "", err
	}

	yamlFilename := yamlFile.Name()
	output, err := this.RunKubectlCommand(command, "-f", yamlFilename)
	removeErr := os.Remove(yamlFilename)

	if err != nil {
		return "", err
	}

	if removeErr != nil {
		return "", removeErr
	}

	return output, err
}
