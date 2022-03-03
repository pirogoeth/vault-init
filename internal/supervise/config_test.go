package supervise

import (
	"testing"
)

func TestConfig(t *testing.T) {
	expectProg := "/bin/ls"
	expectName := "ls"
	expectStr := "/bin/ls -L /directory"

	cfg := &Config{
		Command:       []string{"ls", "-L", "/directory"},
		DisableReaper: false,
	}

	if program, err := cfg.Program(); program != expectProg {
		t.Errorf(
			"expected cfg.Program() to return '%s', got: %s, err: %s",
			expectProg,
			program,
			err,
		)
	}

	if cmdName := cfg.Name(); cmdName != expectName {
		t.Errorf(
			"expected cfg.Name() to return '%s', got: %s",
			expectName,
			cmdName,
		)
	}

	if cmdStr, err := cfg.CommandString(); cmdStr != expectStr {
		t.Errorf(
			"expected cfg.CommandString() to return '%v', got: %v, err: %v",
			expectStr,
			cmdStr,
			err,
		)
	}
}
