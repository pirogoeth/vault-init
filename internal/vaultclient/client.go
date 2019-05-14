package vaultclient

import (
	"context"
	"strings"
	"time"

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

// CreateChildToken creates a token that can be used by the spawned
// child
func (vc *Client) CreateChildToken() (*vaultApi.Secret, error) {
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
	if vc.config.DisableTokenRenew {
		isRenewable = false
	}

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

func (vc *Client) fetchSecrets() ([]*secret, error) {
	logical := vc.vaultClient.Logical()

	secrets := make([]*secret, 0)
	for _, path := range vc.config.Paths {
		secret, err := logical.Read(path)
		if err != nil {
			return nil, errors.Wrapf(err, "could not get secret at path: %s", path)
		}

		secrets = append(secrets, newSecret(vc, path, secret))
	}

	return secrets, nil
}

// intoEnviron templates a set of secret data into a map[string]string to use
// as environment variables to the child program
func (vc *Client) secretsIntoEnviron(secrets []*secret) (map[string]string, error) {
	_, err := vc.secretsAsMap(secrets)
	if err != nil {
		return nil, errors.Wrap(err, "could not create map from secrets")
	}

	return nil, errors.New("Client.secretsIntoEnviron() is not yet implemented")
}

// secretsAsMap merges a bundle of secrets into a single map[string]interface{}
// to be consumed by secretsIntoEnviron.
func (vc *Client) secretsAsMap(secrets []*secret) (map[string]interface{}, error) {
	data := make(map[string]interface{}, 0)
	for _, secret := range secrets {
		pathComponents := strings.Split(secret.Path, "/")
		name := pathComponents[len(pathComponents)-1]
		data[name] = secret.Data
	}

	return data, nil
}

// StartWatcher creates and lanches a watcher that submits environment
// updates to the supervisor.
func (vc *Client) StartWatcher(ctx context.Context, refreshDuration time.Duration) (chan map[string]string, error) {
	// Build an updates channel we can pass back to the supervisor
	updateCh := make(chan map[string]string, 1)

	// Launch the watcher goroutine
	watcher, err := newWatcher(vc, refreshDuration)
	if err != nil {
		return nil, errors.Wrap(err, "while creating watcher")
	}

	go watcher.Watch(ctx, updateCh)

	return updateCh, nil
}
