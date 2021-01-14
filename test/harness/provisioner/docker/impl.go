package docker

import "glow.dev.maio.me/seanj/vault-init/test/harness/provisioner"

var _ provisioner.Provisioner = (*Provisioner)(nil)

func (p *Provisioner) Provision() error {
	return nil
}
