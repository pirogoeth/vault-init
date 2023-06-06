package template

import (
	"text/template"
)

// EnvTemplate holds a reference to a template
type EnvTemplate struct {
	key      string
	value    string
	template *template.Template
}
