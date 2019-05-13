package supervise

// Config holds the configuration for the supervisor
type Config struct {
	// command is the command including executable name/path and arguments
	// that should be spawned
	Command []string

	// disableReaper tells the supervisor not to start the subprocess
	// reaper for cases when vault-init is not running as pid 1
	DisableReaper bool
}

// Supervisor is the actual supervisor instance, providing methods
// to signal, start, and stop the child.
type Supervisor struct {
	// config is the supervisor `Config` object
	config *Config
}
