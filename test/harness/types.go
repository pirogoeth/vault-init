package harness

import (
	"github.com/hashicorp/vault/api"
	"glow.dev.maio.me/seanj/vault-init/initializer"
	"glow.dev.maio.me/seanj/vault-init/test/harness/provisioner"
)

type mountFixture struct {
	Path   string                 `json:"path"`
	Config map[string]interface{} `json:"config"`
}

type secretFixture struct {
	Path string                 `json:"path"`
	Data map[string]interface{} `json:"data"`
}

type fixtures struct {
	Mounts  []*mountFixture  `json:"mounts"`
	Secrets []*secretFixture `json:"secrets"`
}

type vaultConnectionInfo struct {
	Config    *api.Config    `json:"config,omitempty"`
	TLSConfig *api.TLSConfig `json:"tls,omitempty"`
}

type vaultProvider struct {
	// Managed causes the harness to provision a Vault instance via one of the
	// supported provisioner backends.
	Managed bool `json:"managed"`
	// Provisioner provides a common config to be used by any of the supported
	// provisioner backends.
	Provisioner *provisioner.Config `json:"provisioner,omitempty"`
	// Connection is an embed of Vault's API client's api.Config and api.TLSConfig, for
	// configuring a connection to an "unmanaged" Vault (ie., Vault that is NOT managed by this harness)
	Connection *vaultConnectionInfo `json:"connection,omitempty"`
}

type Scenario struct {
	Fixtures      fixtures            `json:"fixtures"`
	VaultInitCfg  *initializer.Config `json:"vault-init"`
	VaultProvider *vaultProvider      `json:"vault"`
	Tests         []string            `json:"tests"`

	filepath string `json:"-"`
}
