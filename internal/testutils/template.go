package testutils

import (
	"bytes"
	"text/template"
)

// RenderTemplate is a helper function to take a template and some configuration and
// return the result as a string.
//
// Where:
//   - tpl is a string like "Hello, {{.Name}}"
//   - data is a struct like `mydata{Name: "World"}`
func RenderTemplate(tpl string, data any) string {
	var result bytes.Buffer

	tmpl, err := template.New("template").Parse(tpl)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(&result, data)
	if err != nil {
		panic(err)
	}

	return result.String()
}
