package podman

import (
	"context"
	"encoding/json"
	"fmt"

	libpodcfg "github.com/containers/common/pkg/config"
	"github.com/containers/podman/v3/libpod"
	vaultApi "github.com/hashicorp/vault/api"

	"github.com/pirogoeth/vault-init/pkg/harness/provisioner"
)

var _ provisioner.Provisioner = (*Provisioner)(nil)

func New() provisioner.Provisioner {
	ctx, cancel := context.WithCancel(context.TODO())
	return &Provisioner{
		ctx:       ctx,
		ctxCancel: cancel,
	}
}

func (p *Provisioner) Configure(cfgJson *json.RawMessage) error {
	cfg := new(Config)
	if err := json.Unmarshal(*cfgJson, cfg); err != nil {
		return fmt.Errorf("while configuring podman provisioner, could not parse provisioner config: %w", err)
	}
	p.cfg = cfg

	var podmanCfg *libpodcfg.Config
	var err error
	if cfg.RuntimeConfig != nil {
		podmanCfg, err = libpodcfg.NewConfig(*cfg.RuntimeConfig)
	} else {
		podmanCfg, err = libpodcfg.Default()
	}
	if err != nil {
		return fmt.Errorf("while configuration podman provisioner, could not load podman user config: %w", err)
	}

	p.runtime, err = libpod.NewRuntimeFromConfig(p.ctx, podmanCfg)
	if err != nil {
		return fmt.Errorf("while configuring podman provisioner, error creating podman runtime: %w", err)
	}

	return nil
}

func (p *Provisioner) Provision() error {
	return nil
}

func (p *Provisioner) Deprovision() error {
	return nil
}

func (p *Provisioner) SpawnVaultAPIClient() (*vaultApi.Client, error) {
	return nil, nil
}
