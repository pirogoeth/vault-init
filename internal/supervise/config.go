package supervise

import (
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
)

// Program returns the path to the program to run
func (c *Config) Program() (string, error) {
	program := c.Command[0]
	if path.IsAbs(program) {
		return path.Clean(program), nil
	}

	program, err := exec.LookPath(program)
	if err != nil {
		log.Errorf("Could not find %s in $PATH", program)
		return "", err
	}

	return program, nil
}

// Args returns the arguments to the program provided
func (c *Config) Args() []string {
	return c.Command[1:]
}

// CommandString returns the command to execute as a string
func (c *Config) CommandString() (string, error) {
	prog, err := c.Program()
	if err != nil {
		return "", errors.Wrap(err, "could not get program path")
	}

	return strings.Join(append([]string{prog}, c.Args()...), " "), nil
}
