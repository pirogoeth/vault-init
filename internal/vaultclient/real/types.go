package real

import (
	vaultApi "github.com/hashicorp/vault/api"

	"glow.dev.maio.me/seanj/vault-init/internal/vaultclient"
	"glow.dev.maio.me/seanj/vault-init/internal/watcher"
)

// Client is a wrapper around the Vault API client
type Client struct {
	config        *vaultclient.Config
	vaultClient   *vaultApi.Client
	tokenRenewer  *vaultApi.Renewer
	secretWatcher *watcher.Watcher
}
