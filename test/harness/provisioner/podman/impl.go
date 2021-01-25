package podman

import (
	"encoding/json"

	vaultApi "github.com/hashicorp/vault/api"

	"glow.dev.maio.me/seanj/vault-init/test/harness/provisioner"
)

var _ provisioner.Provisioner = (*Provisioner)(nil)

func New() provisioner.Provisioner {
	return &Provisioner{}
}

func (p *Provisioner) Configure(cfg *json.RawMessage) error {
	return nil
}

func (p *Provisioner) Provision() error {
	return nil
}

func (p *Provisioner) Deprovision() error {
	return nil
}

func (p *Provisioner) GenerateVaultAPIConfig() (*vaultApi.Config, error) {
	return nil, nil
}
