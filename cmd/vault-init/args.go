package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	vaultApi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	defaultNoInheritToken bool   = false
	defaultNoReaper       bool   = false
	defaultOrphanToken    bool   = false
	defaultTokenRenew     string = "15s"
	defaultTokenTTL       string = ""
	defaultVerbose        bool   = false
)

type args struct {
	Command            []string       `arg:"positional"`
	InitAccessPolicies []string       `arg:"--access-policy,separate,env:INIT_ACCESS_POLICIES" help:"Access policies to create Vault token with"`
	InitOrphanToken    *bool          `arg:"--orphan-token,env:INIT_ORPHAN_TOKEN" help:"Should the created token be independent of the parent"`
	InitNoInheritToken *bool          `arg:"--no-inherit-token,env:INIT_NO_INHERIT_TOKEN" help:"Should the created token be passed down to the spawned child"`
	InitNoReaper       *bool          `arg:"--without-reaper,env:INIT_NO_REAPER" help:"Disable the subprocess reaper"`
	InitPaths          []string       `arg:"--path,separate,env:INIT_PATHS" help:"Secret path to load into template context"`
	InitTokenRenew     *time.Duration `arg:"--token-renewal,env:INIT_TOKEN_RENEWAL" help:"Period at which to renew the Vault token"`
	InitTokenTTL       string         `arg:"--token-ttl,env:INIT_TOKEN_TTL" help:"TTL of the token, minimum duration of 1 hour"`
	VaultAddress       string         `arg:"--vault-address,env:VAULT_ADDR" help:"Address to use to connect to Vault"`
	VaultToken         string         `arg:"--vault-token,env:VAULT_TOKEN" help:"Token to use to authenticate to Vault"`
	VaultTokenFile     string         `arg:"--vault-token-file,env:VAULT_TOKEN_FILE" help:"File containing token to use to authenticate to Vault"`
	Verbose            *bool          `arg:"-v,env:VERBOSE" help:"Enable verbose debug logging"`
}

func (args) Version() string {
	return "vault-init 0.1.0"
}

func (a *args) CheckAndSetDefaults() error {
	var err error

	if a.InitOrphanToken == nil {
		a.InitOrphanToken = new(bool)
		*a.InitOrphanToken = defaultOrphanToken
	}

	if a.InitNoInheritToken == nil {
		a.InitNoInheritToken = new(bool)
		*a.InitNoInheritToken = defaultNoInheritToken
	}

	if a.InitNoReaper == nil {
		a.InitNoReaper = new(bool)
		*a.InitNoReaper = defaultNoReaper
	}

	if a.InitTokenRenew == nil {
		a.InitTokenRenew = new(time.Duration)
		*a.InitTokenRenew, err = time.ParseDuration(defaultTokenRenew)
		if err != nil {
			return errors.Wrapf(err, "could not parse default token renewal duration: `%s`", defaultTokenRenew)
		}
	}

	if a.InitTokenTTL == "" {
		a.InitTokenTTL = defaultTokenTTL
	}

	if a.Verbose == nil {
		a.Verbose = new(bool)
		*a.Verbose = defaultVerbose
	}

	if a.VaultAddress != "" {
		if os.Getenv(vaultApi.EnvVaultAddress) != a.VaultAddress {
			os.Setenv(vaultApi.EnvVaultAddress, a.VaultAddress)
		}
	}

	if a.VaultToken == "" && a.VaultTokenFile == "" {
		return fmt.Errorf("Both VaultToken and VaultTokenFile are unset")
	} else if a.VaultToken != "" && a.VaultTokenFile != "" {
		log.Warnf("Both VaultToken and VaultTokenFile are set, ignoring VaultTokenFile!")
		a.VaultTokenFile = ""
	}

	if a.VaultTokenFile != "" {
		tokenFile, err := os.Open(a.VaultTokenFile)
		if err != nil {
			return errors.Wrap(err, "could not open VaultTokenFile")
		}

		defer tokenFile.Close()

		tokenFileStat, err := tokenFile.Stat()
		if err != nil {
			return errors.Wrap(err, "could not stat() VaultTokenFile")
		}

		tokenFileSize := tokenFileStat.Size()
		tokenBuf := make([]byte, tokenFileSize)
		tokenFile.Read(tokenBuf)

		a.VaultToken = strings.TrimSpace(string(tokenBuf))
	}

	if a.VaultToken != "" {
		if os.Getenv(vaultApi.EnvVaultToken) != a.VaultToken {
			os.Setenv(vaultApi.EnvVaultToken, a.VaultToken)
		}
	}

	return nil
}
