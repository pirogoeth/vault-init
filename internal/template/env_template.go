package template

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
)

// NewEnvTemplate creates an EnvTemplate instance
func NewEnvTemplate(envKey, envValue string) (*EnvTemplate, error) {
	tpl, err := template.New(envKey).Parse(envValue)
	if err != nil {
		log.WithError(err).Errorf("Error while parsing template for environment var: %s", envKey)
		return nil, errors.Wrapf(err, "could not parse template")
	}

	envTpl := &EnvTemplate{
		key:      envKey,
		value:    envValue,
		template: tpl,
	}
	envTpl.template = envTpl.template.Funcs(makeFuncMap())

	return envTpl, nil
}

// Render returns a string with the rendered template
func (e *EnvTemplate) Render(context map[string]interface{}) (string, error) {
	rendered := bytes.NewBufferString("")

	err := e.template.Execute(rendered, context)
	if err != nil {
		return "", errors.Wrap(err, "could not render template")
	}

	return rendered.String(), nil
}
