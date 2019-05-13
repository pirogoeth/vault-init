package vaultclient

import (
	"github.com/davecgh/go-spew/spew"
	vaultApi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// NewClient creates a new Vault API client wrapper
func NewClient(config *Config) (*Client, error) {
	log.WithField("config", config).Debugf("initializing Vault API client")

	vaultClient, err := vaultApi.NewClient(config.Config)
	if err != nil {
		return nil, errors.Wrap(err, "could not create Vault API client")
	}

	return &Client{
		vaultClient: vaultClient,
		config:      config,
	}, nil
}

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

	log.WithFields(log.Fields{
		"initialized": health.Initialized,
		"sealed":      health.Sealed,
		"standby":     health.Standby,
		"version":     health.Version,
		"serverTime":  health.ServerTimeUTC,
	}).Debugf("Vault health status")
	return nil
}

// BuildChildToken creates a token that can be used by the spawned
// child
func (vc *Client) BuildChildToken() (*vaultApi.Secret, error) {
	tokenAuth := vc.vaultClient.Auth().Token()
	var creatorFn TokenCreatorFunc
	var noTokenParent bool

	if vc.config.OrphanToken {
		creatorFn = tokenAuth.CreateOrphan
		noTokenParent = true
	} else {
		creatorFn = tokenAuth.Create
		noTokenParent = false
	}

	isRenewable := true

	secret, err := creatorFn(&vaultApi.TokenCreateRequest{
		Policies:  vc.config.AccessPolicies,
		NoParent:  noTokenParent,
		Renewable: &isRenewable,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not create child token")
	}

	return secret, nil
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

func (vc *Client) collectSecrets() ([]*vaultApi.Secret, error) {
	logical := vc.vaultClient.Logical()

	secrets := make([]*vaultApi.Secret, 0)
	for _, path := range vc.config.Paths {
		secret, err := logical.Read(path)
		if err != nil {
			return nil, errors.Wrapf(err, "could not get secret at path: %s", path)
		}

		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// BuildContext compiles all secret paths into a map that can be used
// as a template context.
func (vc *Client) BuildContext() (map[string]interface{}, error) {
	secrets, err := vc.collectSecrets()
	if err != nil {
		return nil, errors.Wrap(err, "could not collect secrets for context")
	}

	spew.Dump(secrets)

	return nil, nil
}
