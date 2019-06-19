package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	vaultApi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const (
	defaultDebug             bool   = false
	defaultDisableTokenRenew bool   = false
	defaultLogFormat         string = "default"
	defaultNoInheritToken    bool   = false
	defaultNoReaper          bool   = false
	defaultOrphanToken       bool   = false
	defaultRefreshDuration   string = "15s"
	defaultTokenPeriod       string = ""
	defaultTokenTTL          string = ""
	defaultVerbose           bool   = false
)

type args struct {
	Command []string `arg:"positional"`

	AccessPolicies    []string       `arg:"-A,--access-policy,separate,env:INIT_ACCESS_POLICIES" help:"Access policies to create Vault token with"`
	Debug             *bool          `arg:"-D,--debug,env:INIT_DEBUG" help:"Enable super verbose debugging output, which may print sensitive data to terminal"`
	DisableTokenRenew *bool          `arg:"--disable-token-renew,env:INIT_DISABLE_TOKEN_RENEW" help:"Make the child token unable to be renewed"`
	LogFormat         string         `arg:"--log-format,env:INIT_LOG_FORMAT" help:"Change the format used for logging [default, plain, json]"`
	NoInheritToken    *bool          `arg:"--no-inherit-token,env:INIT_NO_INHERIT_TOKEN" help:"Should the created token be passed down to the spawned child"`
	NoReaper          *bool          `arg:"--without-reaper,env:INIT_NO_REAPER" help:"Disable the subprocess reaper"`
	OrphanToken       *bool          `arg:"--orphan-token,env:INIT_ORPHAN_TOKEN" help:"Should the created token be independent of the parent"`
	Paths             []string       `arg:"-p,--path,separate,env:INIT_PATHS" help:"Secret path to load into template context"`
	RefreshDuration   *time.Duration `arg:"--refresh-duration,env:INIT_REFRESH_DURATION" help:"How frequently secrets should be checked for version changes"`

	// TokenPeriod will cause the child token to be created as a periodic token:
	// https://www.vaultproject.io/docs/concepts/tokens.html#periodic-tokens
	TokenPeriod    string `arg:"--token-period,env:INIT_TOKEN_PERIOD" help:"Renewal period of the child token; creates a periodic token"`
	TokenTTL       string `arg:"--token-ttl,env:INIT_TOKEN_TTL" help:"TTL of the token, maximum suffix is hour"`
	VaultAddress   string `arg:"--vault-address,env:VAULT_ADDR" help:"Address to use to connect to Vault"`
	VaultToken     string `arg:"--vault-token,env:VAULT_TOKEN" help:"Token to use to authenticate to Vault"`
	VaultTokenFile string `arg:"--vault-token-file,env:VAULT_TOKEN_FILE" help:"File containing token to use to authenticate to Vault"`
	Verbose        *bool  `arg:"-v,--verbose,env:INIT_VERBOSE" help:"Enable verbose debug logging"`
}

func (args) Version() string {
	return "vault-init 0.1.0"
}

func (a *args) CheckAndSetDefaults() error {
	var err error

	if a.Debug == nil {
		a.Debug = new(bool)
		*a.Debug = defaultDebug
	}

	if a.DisableTokenRenew == nil {
		a.DisableTokenRenew = new(bool)
		*a.DisableTokenRenew = defaultDisableTokenRenew
	}

	if a.OrphanToken == nil {
		a.OrphanToken = new(bool)
		*a.OrphanToken = defaultOrphanToken
	}

	if a.NoInheritToken == nil {
		a.NoInheritToken = new(bool)
		*a.NoInheritToken = defaultNoInheritToken
	}

	if a.NoReaper == nil {
		a.NoReaper = new(bool)
		*a.NoReaper = defaultNoReaper
	}

	if a.RefreshDuration == nil {
		a.RefreshDuration = new(time.Duration)
		*a.RefreshDuration, err = time.ParseDuration(defaultRefreshDuration)
		if err != nil {
			return errors.Wrapf(err, "could not parse default secret refresh duration: `%s`", defaultRefreshDuration)
		}
	}

	if a.TokenPeriod == "" {
		a.TokenPeriod = defaultTokenPeriod
	}

	if a.TokenTTL == "" {
		a.TokenTTL = defaultTokenTTL
	}

	if a.TokenPeriod != "" && a.TokenTTL != "" {
		return errors.New("TokenTTL and TokenPeriod are mutually exclusive; only one may be set")
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

	if a.LogFormat == "" {
		a.LogFormat = defaultLogFormat
	}

	return nil
}
