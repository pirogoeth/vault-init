package dummy

import "glow.dev.maio.me/seanj/vault-init/internal/vaultclient"

var _ vaultclient.VaultClient = (*Client)(nil)

// NewClient creates a new instance of the dummy VaultClient implementation
func NewClient(config *vaultclient.Config) vaultclient.VaultClient {
	return &Client{config: config}
}
