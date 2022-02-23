package kubectl

import (
	"bytes"
	"text/template"
)

type KubeTemplate interface {
	Render() (string, error) // render the template
}

// A general function to render an embedded template filename.
// The data is the text/template current context.
func RenderTextTemplate(filename string, data interface{}) (string, error) {
	tmpl, readError := Embedded.ReadFile(filename)
	if readError != nil {
		return "", readError
	}
	tmplRslt, createErr := template.New(filename).Parse(string(tmpl))
	if createErr != nil {
		return "", createErr
	}
	var rslt bytes.Buffer
	renderErr := tmplRslt.Execute(&rslt, data)
	return rslt.String(), renderErr
}
