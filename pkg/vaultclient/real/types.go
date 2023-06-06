package real

import (
	vaultApi "github.com/hashicorp/vault/api"

	"github.com/pirogoeth/vault-init/pkg/vaultclient"
	"github.com/pirogoeth/vault-init/pkg/watcher"
)

// Client is a wrapper around the Vault API client
type Client struct {
	config        *vaultclient.Config
	vaultClient   *vaultApi.Client
	tokenRenewer  *vaultApi.Renewer
	secretWatcher *watcher.Watcher
}
