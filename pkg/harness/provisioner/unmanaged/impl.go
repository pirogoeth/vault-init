package unmanaged

import (
	"encoding/json"
	"fmt"

	vaultApi "github.com/hashicorp/vault/api"

	"github.com/pirogoeth/vault-init/pkg/harness/provisioner"
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

func (p *Provisioner) SpawnVaultAPIClient() (*vaultApi.Client, error) {
	vaultCfg := vaultApi.DefaultConfig()
	if err := vaultCfg.ReadEnvironment(); err != nil {
		return nil, fmt.Errorf("vault config init failed: %w", err)
	}

	vaultCli, err := vaultApi.NewClient(vaultCfg)
	if err != nil {
		return nil, fmt.Errorf("vault client init failed: %w", err)
	}

	return vaultCli, nil
}
