package dummy

import (
	"context"
	"time"

	vaultApi "github.com/hashicorp/vault/api"

	"github.com/pirogoeth/vault-init/internal/secret"
	"github.com/pirogoeth/vault-init/internal/vaultclient"
)

// Check checks the Vault client's connection and authentication information.
func (vc *Client) Check() error {
	return nil
}

// CreateChildToken creates a token that is a child of the token that the client is connecting with.
// This token is provided to the application that is running underneath vault-init, that way the token
// used by vault-init itself is never exposed to the managed application.
func (vc *Client) CreateChildToken(string) (*vaultApi.Secret, error) {
	return &vaultApi.Secret{}, nil
}

// FetchSecret fetches a secret from Vault, wrapping it into a *secret.Secret.
func (vc *Client) FetchSecret(string) (*secret.Secret, error) {
	return nil, nil
}

// FetchSecrets fetches all of the secrets defined in the configuration.
func (vc *Client) FetchSecrets() ([]*secret.Secret, error) {
	return nil, nil
}

// GetConfig returns the loaded config.
func (vc *Client) GetConfig() *vaultclient.Config {
	return vc.config
}

// InjectChildContext inserts context into the pre-environment data-map
// to provide Vault context, etc, to the child process.
func (vc *Client) InjectChildContext(target map[string]interface{}) (map[string]interface{}, error) {
	return target, nil
}

// NewLeaseRenewer creates a goroutine that constantly renews the secret lease that is configured
// in the *vaultApi.RenewerInput.
func (vc *Client) NewLeaseRenewer(*vaultApi.RenewerInput) (*vaultApi.Renewer, error) {
	return nil, nil
}

// RevokeSecret revokes a leased secret.
func (vc *Client) RevokeSecret(*secret.Secret) error {
	return nil
}

// RevokeTokenAccessor takes a token's accessor and revokes it.
func (vc *Client) RevokeTokenAccessor(string) error {
	return nil
}

// RevokeLease takes a lease ID and revokes it.
func (vc *Client) RevokeLease(string) error {
	return nil
}

// ReadLogical reads the secret at a given logical path inside of Vault.
func (vc *Client) ReadLogical(string) (*vaultApi.Secret, error) {
	return nil, nil
}

// SetToken sets the token that should be used to authenticate to Vault.
func (vc *Client) SetToken(string) error {
	return nil
}

// StartWatcher starts the client's secret watcher. The resulting string channel will receive
// a string array of rendered environment variables when updates happen.
func (vc *Client) StartWatcher(context.Context, time.Duration) (chan []string, error) {
	out := make(chan []string, 1)
	return out, nil
}

// StartSecretRenewer starts a renewer for the given secret.
func (vc *Client) StartSecretRenewer(*secret.Secret) error {
	return nil
}

// StopSecretRenewer stops a renewer for the given secret.
func (vc *Client) StopSecretRenewer(*secret.Secret) error {
	return nil
}
