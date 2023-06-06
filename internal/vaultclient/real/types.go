package real

import (
	vaultApi "github.com/hashicorp/vault/api"

	"github.com/pirogoeth/vault-init/internal/vaultclient"
	"github.com/pirogoeth/vault-init/internal/watcher"
)

// Client is a wrapper around the Vault API client
type Client struct {
	config        *vaultclient.Config
	vaultClient   *vaultApi.Client
	tokenRenewer  *vaultApi.Renewer
	secretWatcher *watcher.Watcher
}
