package real

import (
	"context"
	"time"

	vaultApi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/pirogoeth/vault-init/internal/secret"
	"github.com/pirogoeth/vault-init/internal/vaultclient"
	"github.com/pirogoeth/vault-init/internal/watcher"
)

var _ vaultclient.VaultClient = (*Client)(nil)

// Check performs a healthcheck on the Vault server
func (vc *Client) Check() error {
	sysClient := vc.vaultClient.Sys()
	health, err := sysClient.Health()
	if err != nil {
		return errors.Wrap(err, "could not get Vault health")
	}

	if !health.Initialized || health.Sealed || health.Standby {
		err := errors.Errorf("Vault is not healthy: %v", health)
		log.WithError(err).Errorf("Health check failed")
		return err
	}

	log.WithFields(logrus.Fields{
		"initialized": health.Initialized,
		"sealed":      health.Sealed,
		"standby":     health.Standby,
		"version":     health.Version,
		"serverTime":  health.ServerTimeUTC,
	}).Debugf("Vault health seems ok")
	return nil
}

// CreateChildToken creates a token that can be used by the spawned
// child
func (vc *Client) CreateChildToken(displayName string) (*vaultApi.Secret, error) {
	tokenAuth := vc.vaultClient.Auth().Token()
	var creatorFn vaultclient.TokenCreatorFunc
	var noTokenParent bool

	if vc.config.OrphanToken {
		creatorFn = tokenAuth.CreateOrphan
		noTokenParent = true
	} else {
		creatorFn = tokenAuth.Create
		noTokenParent = false
	}

	renewable := !vc.config.DisableTokenRenew

	createReq := &vaultApi.TokenCreateRequest{
		NoParent:  noTokenParent,
		Policies:  vc.config.AccessPolicies,
		Renewable: &renewable,
	}

	if vc.config.TokenTTL != "" && vc.config.TokenPeriod != "" {
		return nil, errors.New("TokenTTL and TokenPeriod are mutually exclusive; only one may be set")
	}

	if vc.config.TokenTTL != "" {
		createReq.TTL = vc.config.TokenTTL
	}

	if vc.config.TokenPeriod != "" {
		createReq.Period = vc.config.TokenPeriod
	}

	sec, err := creatorFn(createReq)
	if err != nil {
		return nil, errors.Wrap(err, "could not create child token")
	}

	return sec, nil
}

func (vc *Client) GetConfig() *vaultclient.Config {
	return vc.config
}

// SetToken sets the underlying Vault authentication token. Performs a
// `Check()` call to validate authentication.
func (vc *Client) SetToken(v string) error {
	vc.vaultClient.SetToken(v)

	if err := vc.Check(); err != nil {
		return errors.Wrap(err, "could not validate auth with child token")
	}

	return nil
}

// FetchSecret fetches an individual secret path from Vault, wrapping it into a *secret.Secret.
func (vc *Client) FetchSecret(path string) (*secret.Secret, error) {
	sec, err := vc.ReadLogical(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get secret at path: %s", path)
	}

	return secret.New(path, sec), nil
}

// FetchSecrets fetches all the secret paths listed in the configuration.
func (vc *Client) FetchSecrets() ([]*secret.Secret, error) {
	secrets := make([]*secret.Secret, 0)
	for _, path := range vc.config.Paths {
		sec, err := vc.FetchSecret(path)
		if err != nil {
			return nil, errors.Wrapf(err, "could not get secret at path: %s", path)
		}

		if sec == nil {
			log.Warnf("secret at %s is nil, skipping", path)
			continue
		}

		secrets = append(secrets, sec)
	}

	return secrets, nil
}

// InjectChildContext injects configured context into the pre-environment data map.
func (vc *Client) InjectChildContext(dataMap map[string]interface{}) (map[string]interface{}, error) {
	// If token inheritance is enabled, include the Vault connection
	// information in the environment context
	if !vc.config.NoInheritToken {
		dataMap["Vault"] = vc.vaultSettingsAsMap()
	}

	return dataMap, nil
}

// ReadLogical reads the secret at a given logical path inside of Vault.
func (vc *Client) ReadLogical(path string) (*vaultApi.Secret, error) {
	sec, err := vc.vaultClient.Logical().Read(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read secret at path: %s", path)
	}

	return sec, nil
}

// RevokeSecret revokes a leased secret.
func (vc *Client) RevokeSecret(sec *secret.Secret) error {
	if sec.Auth != nil {
		accessor, err := sec.TokenAccessor()
		if err != nil {
			return errors.Wrapf(err, "could not get token accessor for secret loaded from path: %s", sec.Path)
		}

		return vc.RevokeTokenAccessor(accessor)
	} else if sec.LeaseID != "" {
		return vc.RevokeLease(sec.LeaseID)
	} else {
		log.Warnf("Secret at `%s` can not be revoked; it is neither auth nor leased secret", sec.Path)
	}

	return nil
}

// RevokeTokenAccessor takes a token's accessor and revokes it.
func (vc *Client) RevokeTokenAccessor(accessor string) error {
	tokenSys := vc.vaultClient.Auth().Token()
	if err := tokenSys.RevokeAccessor(accessor); err != nil {
		return errors.Wrapf(err, "could not revoke token by accessor: %s", accessor)
	}

	return nil
}

// RevokeLease takes a lease ID and revokes it.
func (vc *Client) RevokeLease(leaseID string) error {
	if err := vc.vaultClient.Sys().Revoke(leaseID); err != nil {
		return errors.Wrapf(err, "could not revoke lease by id: %s", leaseID)
	}

	return nil
}

// NewLeaseRenewer creates a goroutine that constantly renews the secret lease that is configured
// in the *vaultApi.RenewerInput.
func (vc *Client) NewLeaseRenewer(renewerCfg *vaultApi.RenewerInput) (*vaultApi.Renewer, error) {
	renewer, err := vc.vaultClient.NewRenewer(renewerCfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not start secret renewer")
	}

	return renewer, nil
}

// StartWatcher creates and lanches a watcher that submits environment
// updates to the supervisor.
func (vc *Client) StartWatcher(ctx context.Context, refreshDuration time.Duration) (chan []string, error) {
	// Build an updates channel we can pass back to the supervisor
	updateCh := make(chan []string, 1)

	// Launch the watcher goroutine
	watcher, err := watcher.NewWatcher(vc, refreshDuration)
	if err != nil {
		return nil, errors.Wrap(err, "while creating watcher")
	}

	go watcher.Watch(ctx, updateCh)

	return updateCh, nil
}

// StartSecretRenewer starts a renewer for a secret.
func (vc *Client) StartSecretRenewer(sec *secret.Secret) error {
	renewable, err := sec.IsRenewable()
	if err != nil {
		return errors.Wrap(err, "could not check if secret is renewable")
	}

	if !renewable {
		log.Debugf("Secret `%s` is not renewable; skipping renewer", sec.Path)
		return nil
	}

	renewer, err := vc.NewLeaseRenewer(&vaultApi.RenewerInput{
		Secret: sec.Secret,
	})
	if err != nil {
		return errors.Wrap(err, "could not start secret renewer")
	}

	go renewer.Renew()
	go sec.WatchRenewer(renewer)

	return nil
}

// StopSecretRenewer stops a renewer for a secret.
func (vc *Client) StopSecretRenewer(sec *secret.Secret) error {
	renewer := sec.GetRenewer()
	if renewer == nil {
		return errors.Errorf("Asked to stop renewer for secret `%v`, but renewer is nil.", renewer)
	}

	renewer.Stop()

	return nil
}
