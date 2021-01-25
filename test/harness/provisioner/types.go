package provisioner

import (
	"encoding/json"

	vaultApi "github.com/hashicorp/vault/api"
)

// Config is a wrapper around the real configuration that should be loaded by
// the provisioner driver.
type Config struct {
	Driver string          `json:"use"`
	Config json.RawMessage `json:"config"`
}

// Provisioner implementers are expected to create and configure Vault instances
// for use with the test harness.
type Provisioner interface {
	Configure(*json.RawMessage) error
	Provision() error
	Deprovision() error
	GenerateVaultAPIConfig() (*vaultApi.Config, error)
}
