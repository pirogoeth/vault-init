package real

import "glow.dev.maio.me/seanj/vault-init/internal/watcher"

// Client is a wrapper around the Vault API client
type Client struct {
	config        *Config
	vaultClient   *vaultApi.Client
	tokenRenewer  *vaultApi.Renewer
	secretWatcher *watcher.Watcher
}
