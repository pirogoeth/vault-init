package harness

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	vaultApi "github.com/hashicorp/vault/api"

	"glow.dev.maio.me/seanj/vault-init/test/harness/provisioner"
	dockerProv "glow.dev.maio.me/seanj/vault-init/test/harness/provisioner/docker"
	podmanProv "glow.dev.maio.me/seanj/vault-init/test/harness/provisioner/podman"
	unmanagedProv "glow.dev.maio.me/seanj/vault-init/test/harness/provisioner/unmanaged"
)

func (s *Scenario) SetupFixtures(vaultCli *vaultApi.Client) error {
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

	log.Debugf("Engine mounted at %s", mount.Path)

	return nil
}

func (s *Scenario) createSecret(vaultCli *vaultApi.Client, secret *secretFixture) error {
	_, err := vaultCli.Logical().Write(secret.Path, secret.Data)
	if err != nil {
		return fmt.Errorf("during secret fixture %s setup: error while creating secret: %w", secret.Path, err)
	}

	data, err := vaultCli.Logical().Read(secret.Path)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	log.Debugf("Secret created: %#v", data)

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
	if err := vaultCli.Sys().Unmount(mount.Path); err != nil {
		return fmt.Errorf("during mount fixture %s teardown: error while unmounting engine: %w", mount.Path, err)
	}

	log.Debugf("Engine %s unmounted", mount.Path)

	return nil
}

func (s *Scenario) deleteSecret(vaultCli *vaultApi.Client, secret *secretFixture) error {
	if _, err := vaultCli.Logical().Delete(secret.Path); err != nil {
		return fmt.Errorf("during secret fixture %s teardown: error while deleting secret: %w", secret.Path, err)
	}

	log.Debugf("Secret %s deleted", secret.Path)

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
	case "unmanaged":
		provisioner = unmanagedProv.New()
	default:
		return nil, fmt.Errorf("unknown provisioner driver: %s", cfg.Driver)
	}

	if err := provisioner.Configure(&cfg.Config); err != nil {
		return nil, fmt.Errorf("while creating provisioner %s, failed to configure provisioner: %w", cfg.Driver, err)
	}

	return provisioner, nil
}

func (s *Scenario) RunTests() error {
	spew.Dump(s.Tests)
	return fmt.Errorf("not implemented")
}