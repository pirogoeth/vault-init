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
			return nil, fmt.Errorf("%s: %w", path, err)
		}

		buf := bytes.NewBufferString("")
		if _, err := buf.ReadFrom(file); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}

		var scenario Scenario
		if err := yaml.Unmarshal(buf.Bytes(), &scenario); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
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
			provisioner, err := scenario.CreateProvisioner()
			if err != nil {
				return fmt.Errorf("during scenario %s, the provisioner could not be built: %w", scenario.filepath, err)
			}

			if err := provisioner.Provision(); err != nil {
				return fmt.Errorf("during scenario %s, the provisioner failed: %w", scenario.filepath, err)
			}
			defer log.Infof("Deprovisioning result: %#v", provisioner.Deprovision())

			vaultCfg, err = provisioner.GenerateVaultAPIConfig()
			if err != nil {
				return fmt.Errorf(
					"during scenario %s, the provisioner failed to generate a Vault client config: %w",
					scenario.filepath,
					err,
				)
			}

		} else {
			vaultCfg = vaultApi.DefaultConfig()
			if err := vaultCfg.ReadEnvironment(); err != nil {
				return fmt.Errorf("during scenario %s: vault config init failed: %w", scenario.filepath, err)
			}
		}

		vaultCli, err := vaultApi.NewClient(vaultCfg)
		if err != nil {
			return fmt.Errorf("during scenario %s, vault client init failed: %w", scenario.filepath, err)
		}

		if err := scenario.SetupFixtures(vaultCli); err != nil {
			return fmt.Errorf(
				"during scenario %s, the harness failed to set up fixtures: %w",
				scenario.filepath,
				err,
			)
		}

	}
	return nil
}
