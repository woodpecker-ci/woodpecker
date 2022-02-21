package kubectl

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"text/template"
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
	ID        string       // the run id
	Name      string       // the run name
	Namespace string       // the run namespace
	Backend   *KubeBackend // the backend
}

type KubePodConfig struct {
	Name  string
	Image string
}

func (template *KubeTemplateConfig) Render() (string, error) {
	return "", errors.New("Abstract not implemented")
}

func (template *KubeTemplateConfig) GetKubectlCommandArgs() []string {
	commandArgs := []string{}
	if len(template.Namespace) > 0 {
		commandArgs = append(commandArgs, "-n", template.Namespace)
	}
	return commandArgs
}
