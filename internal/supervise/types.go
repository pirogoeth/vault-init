package supervise

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/mitchellh/go-linereader"
)

// state is a container for passing of state during supervisor events
type state struct {
	child       *exec.Cmd
	childCtx    context.Context
	childCancel context.CancelFunc
	parentCtx   context.Context
}

// forwarderv1 takes a stdout and stderr pipe from a child program
// and muxes them both into our logger
type forwarderv1 struct {
	sync.Mutex

	stdoutCh *linereader.Reader
	stderrCh *linereader.Reader

	outWriters []io.Writer
	errWriters []io.Writer

	cancel context.CancelFunc
}

// Config holds the configuration for the supervisor
type Config struct {
	// Command is the command including executable name/path and arguments
	// that should be spawned
	Command []string

	// DisableReaper tells the supervisor not to start the subprocess
	// reaper for cases when vault-init is not running as pid 1
	DisableReaper bool

	// ForwarderStderrWriters
	ForwarderStderrWriters []io.WriteCloser

	// ForwarderStdoutWriters
	ForwarderStdoutWriters []io.WriteCloser

	// OneShot tells the supervisor not to restart the child after it exits
	OneShot bool
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
