package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/sirupsen/logrus"

	"github.com/pirogoeth/vault-init/internal/version"
	"github.com/pirogoeth/vault-init/pkg/harness"
)

type argsT struct {
	// Scenario is a list of scenarios that are to be run
	Scenario []string `arg:"positional"`

	Debug         *bool `arg:"-D,--debug,env:HARNESS_DEBUG" help:"Enable super verbose debugging output"`
	NoDeprovision *bool `arg:"--no-deprovision,env:HARNESS_NO_DEPROVISION" help:"Instructs the harness to skip deprovisioning the Vault instance"`
}

func (argsT) Version() string {
	return version.Version
}

func (t *argsT) ValidateAndSetDefaults() error {
	if t.Debug == nil {
		t.Debug = new(bool)
		*t.Debug = false
	}

	if t.NoDeprovision == nil {
		t.NoDeprovision = new(bool)
		*t.NoDeprovision = false
	}

	return nil
}

func main() {
	args := argsT{}
	arg.MustParse(&args)

	if err := args.ValidateAndSetDefaults(); err != nil {
		fmt.Printf("Error validating options: %#v\n", err)
		os.Exit(1)
	}

	if *args.Debug {
		logrus.SetLevel(logrus.TraceLevel)
	}

	opts := &harness.ScenarioOpts{
		NoDeprovision: *args.NoDeprovision,
	}

	scenarios, err := harness.LoadScenarios(args.Scenario, opts)
	if err != nil {
		fmt.Printf("Could not load scenarios: %+v\n", err)
		os.Exit(1)
	}
	if err := harness.RunScenarios(scenarios); err != nil {
		fmt.Printf("Error while running scenarios: %+v\n", err)
		os.Exit(1)
	}
}
