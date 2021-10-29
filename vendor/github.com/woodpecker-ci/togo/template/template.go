package template

//go:generate togo tmpl -func FuncMap -format text -encode

import (
	"bytes"
	"io"
	"strings"
	"text/template"
	"unicode"
)

// Execute renders the named template and writes to io.Writer wr.
func Execute(wr io.Writer, name string, data interface{}) error {
	buf := new(bytes.Buffer)
	err := T.ExecuteTemplate(buf, name, data)
	if err != nil {
		return err
	}
	src, err := format(buf)
	if err != nil {
		return err
	}
	_, err = io.Copy(wr, src)
	return err
}

// FuncMap provides extra functions for the templates.
var FuncMap = template.FuncMap{
	"substr":   substr,
	"camelize": camelize,
	"hexdump":  hexdump,
}

func substr(s string, i int) string {
	return s[:i]
}

func camelize(kebab string) (camelCase string) {
	isToUpper := false
	for _, runeValue := range kebab {
		if !isCamelCase(runeValue) {
			continue
		}
		if isToUpper {
			camelCase += strings.ToUpper(string(runeValue))
			isToUpper = false
		} else {
			if runeValue == '-' {
				isToUpper = true
			} else {
				camelCase += string(runeValue)
			}
		}
	}
	return
}

func isCamelCase(r rune) bool {
	return r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
