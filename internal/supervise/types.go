package supervise

import (
	"context"
	"os"

	"github.com/mitchellh/go-linereader"
)

// Config holds the configuration for the supervisor
type Config struct {
	// command is the command including executable name/path and arguments
	// that should be spawned
	Command []string

	// DisableReaper tells the supervisor not to start the subprocess
	// reaper for cases when vault-init is not running as pid 1
	DisableReaper bool
}

// forwarder takes a stdout and stderr pipe from a child program
// and muxes them both into our logger
type forwarder struct {
	stdoutCh *linereader.Reader
	stderrCh *linereader.Reader

	cancel context.CancelFunc
}

// Supervisor is the actual supervisor instance, providing methods
// to signal, start, and stop the child.
type Supervisor struct {
	// config is the supervisor `Config` object
	config *Config

	// stateCh is a channel used internally by the supervisor to communicate
	// child state changes during a wait
	stateCh chan *os.ProcessState

	// lastEnv is the last set of environment variables that were rendered
	// by the vaultclient
	lastEnv []string
}
