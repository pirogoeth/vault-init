package initializer

import (
	"fmt"
	"io"
	"os"
	"strings"

	vaultApi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const (
	defaultDebug                     bool   = false
	defaultDisableTokenRenew         bool   = false
	defaultLogFormat                 string = "default"
	defaultNoInheritToken            bool   = false
	defaultNoReaper                  bool   = false
	defaultOneShot                   bool   = false
	defaultOrphanToken               bool   = false
	defaultRefreshDuration           string = "15s"
	defaultTelemetryCollectorGolang  bool   = false
	defaultTelemetryCollectorProcess bool   = false
	defaultTokenPeriod               string = ""
	defaultTokenTTL                  string = ""
	defaultVerbose                   bool   = false
)

// Config is the configuration for `vault-init` as a whole
// and can be populated by an embedding application or populated
// with arguments from the command line and/or environment variables.
type Config struct {
	Command []string `arg:"positional"`

	AccessPolicies    []string `arg:"-A,--access-policy,separate,env:INIT_ACCESS_POLICIES" help:"Access policies to create Vault token with"`
	Debug             *bool    `arg:"-D,--debug,env:INIT_DEBUG" help:"Enable super verbose debugging output, which may print sensitive data to terminal"`
	DisableTokenRenew *bool    `arg:"--disable-token-renew,env:INIT_DISABLE_TOKEN_RENEW" help:"Make the child token unable to be renewed"`
	LogFormat         string   `arg:"--log-format,env:INIT_LOG_FORMAT" help:"Change the format used for logging [default, plain, json]"`
	NoInheritToken    *bool    `arg:"--no-inherit-token,env:INIT_NO_INHERIT_TOKEN" help:"Should the created token be passed down to the spawned child"`
	NoReaper          *bool    `arg:"--without-reaper,env:INIT_NO_REAPER" help:"Disable the subprocess reaper"`
	OneShot           *bool    `arg:"-O,--one-shot,env:INIT_ONE_SHOT" help:"Do not restart when the child process exits"`
	OrphanToken       *bool    `arg:"--orphan-token,env:INIT_ORPHAN_TOKEN" help:"Should the created token be independent of the parent"`
	Paths             []string `arg:"-p,--path,separate,env:INIT_PATHS" help:"Secret path to load into template context"`
	RefreshDuration   string   `arg:"--refresh-duration,env:INIT_REFRESH_DURATION" help:"How frequently secrets should be checked for version changes"`

	// TokenPeriod will cause the child token to be created as a periodic token:
	// https://www.vaultproject.io/docs/concepts/tokens.html#periodic-tokens
	TokenPeriod    string `arg:"--token-period,env:INIT_TOKEN_PERIOD" help:"Renewal period of the child token; creates a periodic token"`
	TokenTTL       string `arg:"--token-ttl,env:INIT_TOKEN_TTL" help:"TTL of the token, maximum suffix is hour"`
	VaultAddress   string `arg:"--vault-address,env:VAULT_ADDR" help:"Address to use to connect to Vault"`
	VaultToken     string `arg:"--vault-token,env:VAULT_TOKEN" help:"Token to use to authenticate to Vault"`
	VaultTokenFile string `arg:"--vault-token-file,env:VAULT_TOKEN_FILE" help:"File containing token to use to authenticate to Vault"`
	Verbose        *bool  `arg:"-v,--verbose,env:INIT_VERBOSE" help:"Enable verbose debug logging"`

	TelemetryAddress          string `arg:"--telemetry-address,env:INIT_TELEMETRY_ADDR" help:"Address to expose Prometheus telemetry on. Disabled if blank."`
	TelemetryCollectorGolang  *bool  `arg:"--use-go-telemetry-collector,env:INIT_TELEMETRY_COLLECTOR_GOLANG" help:"Whether the Golang telemetry collector should be started."`
	TelemetryCollectorProcess *bool  `arg:"--use-process-telemetry-collector,env:INIT_TELEMETRY_COLLECTOR_PROCESS" help:"Whether the process telemetry collector should be started."`

	// ForwarderStdoutWriters allows an external embedder to capture the child's stdout.
	ForwarderStdoutWriters []io.WriteCloser `arg:"-"`
	// ForwarderStderrWriters allows an external embedder to capture the child's stderr.
	ForwarderStderrWriters []io.WriteCloser `arg:"-"`
}

func (c *Config) Clone() (*Config, error) {
	cloned := &Config{}

	copy(cloned.Command, c.Command)
	copy(cloned.AccessPolicies, c.AccessPolicies)
	copy(cloned.Paths, c.Paths)

	cloned.LogFormat = c.LogFormat
	cloned.RefreshDuration = c.RefreshDuration
	cloned.TokenPeriod = c.TokenPeriod
	cloned.TokenTTL = c.TokenTTL
	cloned.VaultAddress = c.VaultAddress
	cloned.VaultToken = c.VaultToken
	cloned.VaultTokenFile = c.VaultTokenFile
	cloned.TelemetryAddress = c.TelemetryAddress

	if c.Debug != nil {
		cloned.Debug = new(bool)
		*cloned.Debug = *c.Debug
	}

	if c.DisableTokenRenew != nil {
		cloned.DisableTokenRenew = new(bool)
		*cloned.DisableTokenRenew = *c.DisableTokenRenew
	}

	if c.NoInheritToken != nil {
		cloned.NoInheritToken = new(bool)
		*cloned.NoInheritToken = *c.NoInheritToken
	}

	if c.NoReaper != nil {
		cloned.NoReaper = new(bool)
		*cloned.NoReaper = *c.NoReaper
	}

	if c.OneShot != nil {
		cloned.OneShot = new(bool)
		*cloned.OneShot = *c.OneShot
	}

	if c.Verbose != nil {
		cloned.Verbose = new(bool)
		*cloned.Verbose = *c.Verbose
	}

	if c.TelemetryCollectorGolang != nil {
		cloned.TelemetryCollectorGolang = new(bool)
		*cloned.TelemetryCollectorGolang = *c.TelemetryCollectorGolang
	}

	if c.TelemetryCollectorProcess != nil {
		cloned.TelemetryCollectorProcess = new(bool)
		*cloned.TelemetryCollectorProcess = *c.TelemetryCollectorProcess
	}

	if err := cloned.ValidateAndSetDefaults(); err != nil {
		return nil, fmt.Errorf("while cloning an initializer.Config, got error: %w", err)
	}

	return cloned, nil
}

// ValidateAndSetDefaults validates the arguments set inside of the
// configuration and fills in certain slots with defaults, if the values
// are unset.
func (c *Config) ValidateAndSetDefaults() error {
	if c.Debug == nil {
		c.Debug = new(bool)
		*c.Debug = defaultDebug
	}

	if c.DisableTokenRenew == nil {
		c.DisableTokenRenew = new(bool)
		*c.DisableTokenRenew = defaultDisableTokenRenew
	}

	if c.OneShot == nil {
		c.OneShot = new(bool)
		*c.OneShot = defaultOneShot
	}

	if c.OrphanToken == nil {
		c.OrphanToken = new(bool)
		*c.OrphanToken = defaultOrphanToken
	}

	if c.NoInheritToken == nil {
		c.NoInheritToken = new(bool)
		*c.NoInheritToken = defaultNoInheritToken
	}

	if c.NoReaper == nil {
		c.NoReaper = new(bool)
		*c.NoReaper = defaultNoReaper
	}

	if c.RefreshDuration == "" {
		c.RefreshDuration = defaultRefreshDuration
	}

	if c.TokenPeriod == "" {
		c.TokenPeriod = defaultTokenPeriod
	}

	if c.TokenTTL == "" {
		c.TokenTTL = defaultTokenTTL
	}

	if c.Verbose == nil {
		c.Verbose = new(bool)
		*c.Verbose = defaultVerbose
	}

	if c.VaultAddress != "" {
		if os.Getenv(vaultApi.EnvVaultAddress) != c.VaultAddress {
			os.Setenv(vaultApi.EnvVaultAddress, c.VaultAddress)
		}
	}

	if c.TokenPeriod != "" && c.TokenTTL != "" {
		return errors.New("TokenTTL and TokenPeriod are mutually exclusive; only one may be set")
	}

	if c.VaultToken == "" && c.VaultTokenFile == "" {
		return fmt.Errorf("Both VaultToken and VaultTokenFile are unset")
	} else if c.VaultToken != "" && c.VaultTokenFile != "" {
		log.Warnf("Both VaultToken and VaultTokenFile are set, ignoring VaultTokenFile!")
		c.VaultTokenFile = ""
	}

	if c.VaultTokenFile != "" {
		tokenFile, err := os.Open(c.VaultTokenFile)
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

		c.VaultToken = strings.TrimSpace(string(tokenBuf))
	}

	if c.VaultToken != "" {
		if os.Getenv(vaultApi.EnvVaultToken) != c.VaultToken {
			os.Setenv(vaultApi.EnvVaultToken, c.VaultToken)
		}
	}

	if c.LogFormat == "" {
		c.LogFormat = defaultLogFormat
	}

	if c.TelemetryCollectorGolang == nil {
		c.TelemetryCollectorGolang = new(bool)
		*c.TelemetryCollectorGolang = defaultTelemetryCollectorGolang
	}

	if c.TelemetryCollectorProcess == nil {
		c.TelemetryCollectorProcess = new(bool)
		*c.TelemetryCollectorProcess = defaultTelemetryCollectorProcess
	}

	return nil
}
