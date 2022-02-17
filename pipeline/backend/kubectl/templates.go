package kubectl

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type KubeTemplate interface {
	Render() (string, error) // render the template
}

func toKuberenetesValidName(name string, maxChars int) string {
	name = strings.ToLower(name)

	// cleanup chars
	re, _ := regexp.Compile("[^a-z0-9]+")
	name = string(re.ReplaceAll([]byte(name), []byte("-")))

	if len(name) > maxChars {
		name = name[len(name)-maxChars:]
	}

	// cleanup starters and enders
	re, _ = regexp.Compile("^-+|-+$")
	name = string(re.ReplaceAll([]byte(name), []byte("")))

	return name
}

func renderTemplate(name string, data interface{}) (string, error) {
	tmpl, readError := Embedded.ReadFile(name)
	if readError != nil {
		return "", readError
	}
	tmplRslt, createErr := template.New(name).Parse(string(tmpl))
	if createErr != nil {
		return "", createErr
	}
	var rslt bytes.Buffer
	renderErr := tmplRslt.Execute(&rslt, data)
	return rslt.String(), renderErr
}

type KubeTemplateConfig struct {
	ID        string          // the run id
	Name      string          // the run name
	Namespace string          // the run namespace
	Backend   *KubeCtlBackend // the backend
}

type KubePodConfig struct {
	Name  string
	Image string
}

func (t *KubeTemplateConfig) Render() (string, error) {
	return "", errors.New("Abstract not implemented")
}

func (t *KubeTemplateConfig) GetKubectlCommandArgs() []string {
	commandArgs := []string{}
	if len(t.Namespace) > 0 {
		commandArgs = append(commandArgs, "-n", t.Namespace)
	}
	return commandArgs
}

type KubeVolumeTemplate struct {
	StorageClassName string
	StorageSize      string
	Name             string
	Engine           *KubeCtlBackend // the executing engine
}

func (t *KubeVolumeTemplate) VolumeName() string {
	return toKuberenetesValidName(t.Engine.ID()+"-"+t.Name, 60)
}

func (t *KubeVolumeTemplate) Render() (string, error) {
	return renderTemplate("templates/volume_claim.yaml", t)
}

type KubeJobTemplateEnv struct {
	Key   string
	Value string
}

type KubeJobTemplate struct {
	Step   *types.Step     // The executing step
	Engine *KubeCtlBackend // the executing engine
}

func (t *KubeJobTemplate) Render() (string, error) {
	return renderTemplate("templates/step_job.yaml", t)
}

func (t *KubeJobTemplate) JobName() string {
	return toKuberenetesValidName(t.Engine.ID()+"-"+t.Step.Name, 60)
}

func (t *KubeJobTemplate) JobID() string {
	return t.Engine.RunID + "-" + t.Step.Name
}

func (t *KubeJobTemplate) HasEnvironmentVariables() bool {
	return len(t.Step.Environment) > 0
}

func (t *KubeJobTemplate) EnvironmentVariables() []KubeJobTemplateEnv {
	arr := []KubeJobTemplateEnv{}
	for k, v := range t.Step.Environment {
		arr = append(arr, KubeJobTemplateEnv{
			Key:   k,
			Value: fmt.Sprint(v) + "",
		})
	}
	return arr
}

func (t *KubeJobTemplate) ShellCommand() string {
	return strings.Join(t.Step.Command, ";")
}
