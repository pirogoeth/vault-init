package template

import (
	"encoding/json"
	"text/template"
)

func makeFuncMap() template.FuncMap {
	return template.FuncMap{
		"json": encodeAsJSON,
	}
}

func encodeAsJSON(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}
