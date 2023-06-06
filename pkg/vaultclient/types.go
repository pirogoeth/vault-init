package vaultclient

import (
	"context"
	"time"

	vaultApi "github.com/hashicorp/vault/api"
	"github.com/pirogoeth/vault-init/pkg/secret"
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

	// TokenPeriod sets the renewal period of the token. Setting this
	// option will make the child token be a periodic token, which
	// requires a root/sudo token
	TokenPeriod string

	// TokenTTL defaults the lifetime of the token
	TokenTTL string
}

// VaultClient is an interface defining the functions required to be
// implemented to handle tokens
type VaultClient interface {
	// Check checks the Vault client's connection and authentication information.
	Check() error
	// CreateChildToken creates a token that is a child of the token that the client is connecting with.
	// This token is provided to the application that is running underneath vault-init, that way the token
	// used by vault-init itself is never exposed to the managed application.
	CreateChildToken(string) (*vaultApi.Secret, error)
	// FetchSecret fetches a secret from Vault, wrapping it into a *secret.Secret.
	FetchSecret(string) (*secret.Secret, error)
	// FetchSecrets fetches all of the secrets defined in the configuration.
	FetchSecrets() ([]*secret.Secret, error)
	// GetConfig returns the loaded config.
	GetConfig() *Config
	// InjectChildContext inserts context into the pre-environment data-map
	// to provide Vault context, etc, to the child process.
	InjectChildContext(map[string]interface{}) (map[string]interface{}, error)
	// NewLeaseRenewer creates a goroutine that constantly renews the secret lease that is configured
	// in the *vaultApi.RenewerInput.
	NewLeaseRenewer(*vaultApi.RenewerInput) (*vaultApi.Renewer, error)
	// RevokeSecret revokes a leased secret.
	RevokeSecret(*secret.Secret) error
	// RevokeTokenAccessor takes a token's accessor and revokes it.
	RevokeTokenAccessor(string) error
	// RevokeLease takes a lease ID and revokes it.
	RevokeLease(string) error
	// ReadLogical reads the secret at a given logical path inside of Vault.
	ReadLogical(string) (*vaultApi.Secret, error)
	// SetToken sets the token that should be used to authenticate to Vault.
	SetToken(string) error
	// StartWatcher starts the client's secret watcher. The resulting string channel will receive
	// a string array of rendered environment variables when updates happen.
	StartWatcher(context.Context, time.Duration) (chan []string, error)
	// StartSecretRenewer starts a renewer for the given secret.
	StartSecretRenewer(*secret.Secret) error
	// StopSecretRenewer stops a renewer for the given secret.
	StopSecretRenewer(*secret.Secret) error
}
