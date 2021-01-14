package harness

import (
	"bytes"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	vaultApi "github.com/hashicorp/vault/api"
)

func LoadScenarios(scenarioPaths []string) ([]*Scenario, error) {
	scenarios := make([]*Scenario, 0)
	for _, path := range scenarioPaths {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		buf := bytes.NewBufferString("")
		if _, err := buf.ReadFrom(file); err != nil {
			return nil, err
		}

		var scenario Scenario
		if err := yaml.Unmarshal(buf.Bytes(), &scenario); err != nil {
			return nil, err
		}

		// Attach scenario metadata
		scenario.filepath = path

		scenarios = append(scenarios, &scenario)
	}

	return scenarios, nil
}

func RunScenarios(scenarios []*Scenario) error {
	for _, scenario := range scenarios {
		var vaultCfg *vaultApi.Config

		if !scenario.VaultProvider.Managed {
			provisioner, err := scenario.MakeProvisioner()
			if err != nil {
				return fmt.Errorf("during scenario %s, the provisioner could not be built: %w", scenario.filepath, err)
			}

			if err := provisioner.Provision(); err != nil {
				return fmt.Errorf("during scenario %s, the provisioner failed: %w", scenario.filepath, err)
			}

			// XXX - Provision a Vault client against the newly provisioned instance
		} else {
			vaultCfg = vaultApi.DefaultConfig()
			if err := vaultCfg.ReadEnvironment(); err != nil {
				return fmt.Errorf("during vault config init, environment read failed: %w", err)
			}
		}
	}
	return nil
}
