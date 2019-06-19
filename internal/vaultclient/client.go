package vaultclient

import (
	"context"
	"os"
	"strings"
	"time"

	vaultApi "github.com/hashicorp/vault/api"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"glow.dev.maio.me/seanj/vault-init/internal/template"
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
func (vc *Client) CreateChildToken(displayName string) (*secret, error) {
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

	secret, err := creatorFn(createReq)
	if err != nil {
		return nil, errors.Wrap(err, "could not create child token")
	}

	return newSecret(vc, "CHILD_TOKEN", secret), nil
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

		if secret == nil {
			log.Warnf("secret at %s is nil, skipping", path)
			continue
		}

		secrets = append(secrets, newSecret(vc, path, secret))
	}

	return secrets, nil
}

// intoEnviron templates a set of secret data into a map[string]string to use
// as environment variables to the child program
func (vc *Client) secretsIntoEnviron(secrets []*secret) (map[string]string, error) {
	secretCtx, err := vc.secretsAsMap(secrets)
	if err != nil {
		return nil, errors.Wrap(err, "could not create map from secrets")
	}

	environ := os.Environ()
	envMap := make(map[string]string, 0)

	for _, envVar := range environ {
		pair := strings.SplitN(envVar, "=", 2)

		key, value := pair[0], pair[1]
		if vc.keyIsFiltered(key) {
			continue
		}

		tpl, err := template.NewEnvTemplate(key, value)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse environment variable template")
		}

		envMap[key], err = tpl.Render(secretCtx)
		if err != nil {
			return nil, errors.Wrap(err, "could not render environment variable template")
		}
	}

	return envMap, nil
}

func (vc *Client) keyIsFiltered(key string) bool {
	if strings.HasPrefix(key, "INIT_") {
		return true
	} else if strings.HasPrefix(key, "VAULT_") {
		if vc.config.NoInheritToken {
			return true
		}
	}

	return false
}

// secretsAsMap merges a bundle of secrets into a single map[string]interface{}
// to be consumed by secretsIntoEnviron.
func (vc *Client) secretsAsMap(secrets []*secret) (map[string]interface{}, error) {
	data := make(map[string]interface{}, 0)

	for _, secret := range secrets {
		if err := mergo.Merge(&data, secret.dataMap()); err != nil {
			return nil, errors.Wrap(err, "could not merge secret to data")
		}
	}

	// If token inheritance is enabled, include the Vault connection
	// information in the environment context
	if !vc.config.NoInheritToken {
		data["Vault"] = vc.vaultSettingsAsMap()
	}

	return data, nil
}

func (vc *Client) vaultSettingsAsMap() map[string]interface{} {
	data := make(map[string]interface{}, 0)
	tlsConfig := make(map[string]interface{}, 0)

	if v := os.Getenv(vaultApi.EnvVaultCACert); v != "" {
		tlsConfig["ca_cert"] = v
	}

	if v := os.Getenv(vaultApi.EnvVaultCAPath); v != "" {
		tlsConfig["ca_path"] = v
	}

	if v := os.Getenv(vaultApi.EnvVaultClientCert); v != "" {
		tlsConfig["cert"] = v
	}

	if v := os.Getenv(vaultApi.EnvVaultClientKey); v != "" {
		tlsConfig["key"] = v
	}

	if v := os.Getenv(vaultApi.EnvVaultSkipVerify); v != "" {
		tlsConfig["skip_verify"] = v
	}

	if v := os.Getenv(vaultApi.EnvVaultTLSServerName); v != "" {
		tlsConfig["server_name"] = v
	}

	data["address"] = vc.config.Address
	data["agent_address"] = vc.config.AgentAddress
	data["max_retries"] = vc.config.MaxRetries
	data["timeout"] = vc.config.Timeout.String()
	data["tls"] = tlsConfig
	data["token"] = vc.vaultClient.Token()

	return data
}

// StartWatcher creates and lanches a watcher that submits environment
// updates to the supervisor.
func (vc *Client) StartWatcher(ctx context.Context, refreshDuration time.Duration) (chan []string, error) {
	// Build an updates channel we can pass back to the supervisor
	updateCh := make(chan []string, 1)

	// Launch the watcher goroutine
	watcher, err := newWatcher(vc, refreshDuration)
	if err != nil {
		return nil, errors.Wrap(err, "while creating watcher")
	}

	go watcher.Watch(ctx, updateCh)

	return updateCh, nil
}
