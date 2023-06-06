package template

import (
	"encoding/json"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/pirogoeth/vault-init/pkg/vaultclient"
)

func makeFuncMap() template.FuncMap {
	return template.FuncMap{
		"json": encodeAsJSON,
	}
}

func encodeAsJSON(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

// RenderEnvironmentWithDataMap renders an environment variable mapping from
// the data map derived from secrets.
func RenderEnvironmentFromDataMap(cfg *vaultclient.Config, dataMap map[string]interface{}) (map[string]string, error) {
	environ := os.Environ()
	envMap := make(map[string]string, 0)

	for _, envVar := range environ {
		pair := strings.SplitN(envVar, "=", 2)

		key, value := pair[0], pair[1]
		if IsKeyFiltered(cfg, key) {
			continue
		}

		tpl, err := NewEnvTemplate(key, value)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse environment variable template")
		}

		envMap[key], err = tpl.Render(dataMap)
		if err != nil {
			return nil, errors.Wrap(err, "could not render environment variable template")
		}
	}

	return envMap, nil
}

func IsKeyFiltered(cfg *vaultclient.Config, key string) bool {
	if strings.HasPrefix(key, "INIT_") {
		return true
	} else if strings.HasPrefix(key, "VAULT_") {
		if cfg.NoInheritToken {
			return true
		}
	}

	return false
}
