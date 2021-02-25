package harness

import (
	"fmt"
	"strings"

	vaultApi "github.com/hashicorp/vault/api"

	"glow.dev.maio.me/seanj/vault-init/test/harness/provisioner"
	dockerProv "glow.dev.maio.me/seanj/vault-init/test/harness/provisioner/docker"
	podmanProv "glow.dev.maio.me/seanj/vault-init/test/harness/provisioner/podman"
)

func (s *Scenario) SetupFixtures(vaultCli *vaultApi.Client) error {
	log.Infof("fixures %#v", s.Fixtures)
	for _, mount := range s.Fixtures.Mounts {
		if err := s.createMount(vaultCli, mount); err != nil {
			return fmt.Errorf("error while creating mount fixtures: %w", err)
		}
	}

	for _, secret := range s.Fixtures.Secrets {
		if err := s.createSecret(vaultCli, secret); err != nil {
			return fmt.Errorf("error while creating secret fixtures: %w", err)
		}
	}

	return nil
}

func (s *Scenario) createMount(vaultCli *vaultApi.Client, mount *mountFixture) error {
	if err := vaultCli.Sys().Mount(mount.Path, mount.Config); err != nil {
		return fmt.Errorf("during mount fixture %s setup: error while mounting engine: %w", mount.Path, err)
	}

	log.Infof("Engine mounted at %s", mount.Path)

	return nil
}

func (s *Scenario) createSecret(vaultCli *vaultApi.Client, secret *secretFixture) error {
	return nil
}

func (s *Scenario) TeardownFixtures(vaultCli *vaultApi.Client) error {
	for _, secret := range s.Fixtures.Secrets {
		if err := s.deleteSecret(vaultCli, secret); err != nil {
			return fmt.Errorf("error while destroying secret fixtures: %w", err)
		}
	}

	for _, mount := range s.Fixtures.Mounts {
		if err := s.deleteMount(vaultCli, mount); err != nil {
			return fmt.Errorf("error while destroying mount fixtures: %w", err)
		}
	}

	return nil
}

func (s *Scenario) deleteMount(vaultCli *vaultApi.Client, mount *mountFixture) error {
	return nil
}

func (s *Scenario) deleteSecret(vaultCli *vaultApi.Client, secret *secretFixture) error {
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
