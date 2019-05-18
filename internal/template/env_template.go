package template

import (
	"text/template"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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

	return envTpl, nil
}

func (e *EnvTemplate) Render(context map[string]string) {

}
