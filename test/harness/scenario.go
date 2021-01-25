package harness

import (
	"fmt"
	"strings"

	"glow.dev.maio.me/seanj/vault-init/test/harness/provisioner"
	dockerProv "glow.dev.maio.me/seanj/vault-init/test/harness/provisioner/docker"
	podmanProv "glow.dev.maio.me/seanj/vault-init/test/harness/provisioner/podman"
)

func (s *Scenario) SetupFixtures() error {
	for _, mount := range s.Fixtures.Mounts {
		if err := s.createMount(mount); err != nil {
			return fmt.Errorf("error while creating mount fixtures: %w", err)
		}
	}

	for _, secret := range s.Fixtures.Secrets {
		if err := s.createSecret(secret); err != nil {
			return fmt.Errorf("error while creating secret fixtures: %w", err)
		}
	}

	return nil
}

func (s *Scenario) createMount(mount *mountFixture) error {
	return nil
}

func (s *Scenario) createSecret(secret *secretFixture) error {
	return nil
}

func (s *Scenario) TeardownFixtures() error {
	for _, secret := range s.Fixtures.Secrets {
		if err := s.deleteSecret(secret); err != nil {
			return fmt.Errorf("error while destroying secret fixtures: %w", err)
		}
	}

	for _, mount := range s.Fixtures.Mounts {
		if err := s.deleteMount(mount); err != nil {
			return fmt.Errorf("error while destroying mount fixtures: %w", err)
		}
	}

	return nil
}

func (s *Scenario) deleteMount(mount *mountFixture) error {
	return nil
}

func (s *Scenario) deleteSecret(secret *secretFixture) error {
	return nil
}

func (s *Scenario) CreateProvisioner() (provisioner.Provisioner, error) {
	cfg := s.VaultProvider.ProvisionerCfg
	var provisioner provisioner.Provisioner
	switch strings.ToLower(cfg.Driver) {
	case "docker":
		provisioner = dockerProv.New()
	case "podman":
		provisioner = podmanProv.New()
	default:
		return nil, fmt.Errorf("unknown provisioner driver: %s", cfg.Driver)
	}

	if err := provisioner.Configure(&cfg.Config); err != nil {
		return nil, fmt.Errorf("while creating provisioner %s, failed to configure provisioner: %w", cfg.Driver, err)
	}

	return provisioner, nil
}
