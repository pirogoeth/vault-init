package vaultclient

import (
	vaultApi "github.com/hashicorp/vault/api"
)

// NewConfigWithDefaults creates a vaultclient.Config with the
// Vault client's defaults set in the embedded `vaultApi.Config`
func NewConfigWithDefaults() *Config {
	defaults := vaultApi.DefaultConfig()
	return &Config{
		Config: defaults,
	}
}

// ReadEnvironment makes a ReadEnvironment call to the embedded
// `vaultApi.Config` struct
func (c *Config) ReadEnvironment() error {
	return c.Config.ReadEnvironment()
}
