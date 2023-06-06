package harness

import (
	"io"
	"time"

	vaultApi "github.com/hashicorp/vault/api"
	"github.com/pirogoeth/vault-init/initializer"
	"github.com/pirogoeth/vault-init/pkg/harness/provisioner"
)

type mountFixture struct {
	Path   string               `json:"path"`
	Config *vaultApi.MountInput `json:"config"`
}

type secretFixture struct {
	Path string                 `json:"path"`
	Data map[string]interface{} `json:"data"`
}

type fixtures struct {
	Mounts  []*mountFixture  `json:"mounts"`
	Secrets []*secretFixture `json:"secrets"`
}

type vaultConnectionInfo struct {
	Config    *vaultApi.Config    `json:"config,omitempty"`
	TLSConfig *vaultApi.TLSConfig `json:"tls,omitempty"`
}

type vaultHealthcheckConfig struct {
	// Tries is the number of times to attempt the Vault healthcheck
	Tries uint16 `json:"tries,omitempty"`
	// Interval is the wait time between attempts of the Vault healthcheck
	Interval time.Duration `json:"duration,omitempty"`
}

type vaultProvider struct {
	// Managed causes the harness to provision a Vault instance via one of the
	// supported provisioner backends. When unmanaged, the Vault API client's
	// configuration is loaded from the environment. Otherwise, the config is
	// generated by the provisioner.
	Managed bool `json:"managed"`
	// ProvisionerCfg provides a common config to be used by any of the supported
	// provisioner backends.
	ProvisionerCfg *provisioner.Config `json:"provisioner,omitempty"`
	// HealthcheckCfg is the configuration for testing liveness of the provisioned Vault instance
	HealthcheckCfg *vaultHealthcheckConfig `json:"liveness_check,omitempty"`
}

type testSuite struct {
	// Environment sets the expected environment variables for the test suite process.
	Environment map[string]string `json:"env,omitempty"`
	// Suite defines which test suite file should be run.
	Suite string `json:"suite"`
	// Args defines an optional list of command line args to pass to `go test`
	Args []string `json:"args,omitempty"`
}

type testSuiteResult struct {
	// Error is the error returned by the vault-init initializer, if any.
	Error error
	// StderrReader is an io.Reader over the test suite's stderr.
	StderrReader io.ReadCloser
	// StdoutReader is an io.Reader over the test suite's stdout.
	StdoutReader io.ReadCloser
}

// goTestEvent is pulled from the `TestEvent` defined at `go doc test2json`.
// We use this to parse the json output of go test and construct the testSuiteResult structure.
type goTestEvent struct {
	Time    time.Time `json:",omitempty"`
	Action  string
	Package string  `json:",omitempty"`
	Test    string  `json:",omitempty"`
	Elapsed float64 `json:",omitempty"`
	Output  string  `json:",omitempty"`
}

type Scenario struct {
	Fixtures      fixtures            `json:"fixtures"`
	VaultInitCfg  *initializer.Config `json:"vault-init"`
	VaultProvider *vaultProvider      `json:"vault"`
	Tests         []*testSuite        `json:"tests"`

	filepath string        `json:"-"`
	opts     *ScenarioOpts `json:"-"`
}

type ScenarioOpts struct {
	// NoDeprovision instructs the harness not to deprovision the Vault instance.
	NoDeprovision bool
}