package real

import (
	"os"

	vaultApi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"

	"github.com/pirogoeth/vault-init/internal/vaultclient"
)

// NewClient creates a new Vault API client wrapper
func NewClient(config *vaultclient.Config) (vaultclient.VaultClient, error) {
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
