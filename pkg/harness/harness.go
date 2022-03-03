package harness

import (
	"bytes"
	"fmt"
	"os"

	"github.com/ghodss/yaml"

	"glow.dev.maio.me/seanj/vault-init/pkg/harness/provisioner"
)

func LoadScenarios(scenarioPaths []string, opts *ScenarioOpts) ([]*Scenario, error) {
	if opts == nil {
		opts = &ScenarioOpts{}
	}

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
		scenario.opts = opts

		if !scenario.VaultProvider.Managed {
			scenario.VaultProvider.ProvisionerCfg = &provisioner.Config{Driver: "unmanaged"}
		}

		scenarios = append(scenarios, &scenario)
	}

	return scenarios, nil
}

func RunScenarios(scenarios []*Scenario) error {
	for _, scenario := range scenarios {
		provisioner, err := scenario.CreateProvisioner()
		if err != nil {
			return fmt.Errorf("during scenario %s, the provisioner could not be built: %w", scenario.filepath, err)
		}

		if err := provisioner.Provision(); err != nil {
			return fmt.Errorf("during scenario %s, the provisioner failed: %w", scenario.filepath, err)
		}

		defer Teardown(scenario, provisioner)

		vaultCli, err := provisioner.SpawnVaultAPIClient()
		if err != nil {
			return fmt.Errorf(
				"during scenario %s, the provisioner failed to spawn a Vault client: %w",
				scenario.filepath,
				err,
			)
		}

		if err := scenario.SetupFixtures(vaultCli); err != nil {
			return fmt.Errorf(
				"during scenario %s, the harness failed to set up fixtures: %w",
				scenario.filepath,
				err,
			)
		}

		if err := scenario.ConfigureVaultInitFromVaultClient(vaultCli); err != nil {
			return fmt.Errorf(
				"during scenario %s, the harness failed to configure vault-init: %w",
				scenario.filepath,
				err,
			)
		}

		if err := scenario.RunTests(); err != nil {
			return fmt.Errorf(
				"during scenario %s, the harness caught a failure during tests: %w",
				scenario.filepath,
				err,
			)
		}

		if err := scenario.TeardownFixtures(vaultCli); err != nil {
			return fmt.Errorf(
				"during scenario %s, the harness failed to tear down fixtures: %w",
				scenario.filepath,
				err,
			)
		}
	}
	return nil
}

func Teardown(scenario *Scenario, provisioner provisioner.Provisioner) {
	if !scenario.opts.NoDeprovision {
		log.Infof("Deprovisioning result: %#v", provisioner.Deprovision())
	} else {
		log.Infof("Skipping deprovisioning for %s", scenario.filepath)
	}
}
