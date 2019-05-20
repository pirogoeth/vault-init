package vaultclient

import (
	"time"

	vaultApi "github.com/hashicorp/vault/api"
)

// TokenCreatorFunc is a function that returns a token that can be used by
// the child process and by the vault-init
type TokenCreatorFunc func(*vaultApi.TokenCreateRequest) (*vaultApi.Secret, error)

// Config configures the Vault client's operations
type Config struct {
	*vaultApi.Config

	// AccessPolicies is a list of policies the child's Vault token
	// should be created with.
	AccessPolicies []string

	// DisableTokenRenew defines the "renewability" of the token. If true,
	// sets the `renewable` flag to false on token creation.
	DisableTokenRenew bool

	// NoInheritToken controls whether the vaultclient sends VAULT_TOKEN
	// and Vault settings to the child process
	NoInheritToken bool

	// OrphanToken defines whether the created token should be an orphan
	// or not.
	OrphanToken bool

	// Paths is a list of paths that should be inserted into the template
	// context from Vault.
	Paths []string

	// TokenRenew defines how frequently the child's Vault token should be
	// renewed
	TokenRenew time.Duration

	// TokenTTL defaults the lifetime of the token
	TokenTTL string
}

// Client is a wrapper around the Vault API client
type Client struct {
	config        *Config
	vaultClient   *vaultApi.Client
	tokenRenewer  *vaultApi.Renewer
	secretWatcher *watcher
}

type watcher struct {
	client          *Client
	refreshDuration time.Duration
}

type secret struct {
	*vaultApi.Secret

	// client is a reference to the vaultclient.Client
	client *Client

	// renewer is a reference to the running vaultApi.Renewer
	renewer *vaultApi.Renewer

	// Path is the logical path at which this secret was found
	Path string
}
