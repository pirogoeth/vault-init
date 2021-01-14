package harness

import (
	"fmt"

	"glow.dev.maio.me/seanj/vault-init/test/harness/provisioner"
)

func (s *Scenario) SetupFixtures() error {
	for _, mount := range s.Fixtures.Mounts {
		if err := s.createMount(mount); err != nil {
			return fmt.Errorf("error while creating mount fixtures", err)
		}
	}

	for _, secret := range s.Fixtures.Secrets {
		if err := s.createSecret(secret); err != nil {
			return fmt.Errorf("error while creating secret fixtures", err)
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
			return fmt.Errorf("error while creating secret fixtures", err)
		}
	}

	for _, mount := range s.Fixtures.Mounts {
		if err := s.deleteMount(mount); err != nil {
			return fmt.Errorf("error while creating mount fixtures", err)
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

func (s *Scenario) MakeProvisioner() (provisioner.Provisioner, error) {
	return nil, nil
}
