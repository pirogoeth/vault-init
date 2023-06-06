package dummy

import "github.com/pirogoeth/vault-init/internal/vaultclient"

var _ vaultclient.VaultClient = (*Client)(nil)

// NewClient creates a new instance of the dummy VaultClient implementation
func NewClient(config *vaultclient.Config) (vaultclient.VaultClient, error) {
	return &Client{config: config}, nil
}
