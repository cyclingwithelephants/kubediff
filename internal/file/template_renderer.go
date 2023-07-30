package file

import (
	"bytes"
	"text/template"
)

type TemplateRenderer struct{}

func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{}
}

func (T *TemplateRenderer) Render(templateString string, replacements map[string]string) (string, error) {
	parsedTemplate, err := template.New("template").Parse(templateString)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = parsedTemplate.Execute(
		&buf,
		replacements,
	)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
