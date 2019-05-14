package supervise

import (
	"os/exec"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
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
func (c *Config) CommandString() string {
	return strings.Join(c.Command, " ")
}
