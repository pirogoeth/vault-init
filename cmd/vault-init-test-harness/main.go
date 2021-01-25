package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"

	"glow.dev.maio.me/seanj/vault-init/internal/version"
	"glow.dev.maio.me/seanj/vault-init/test/harness"
)

type argsT struct {
	// Scenario is a list of scenarios that are to be run
	Scenario []string `arg:"positional"`
}

func (argsT) Version() string {
	return version.Version
}

func main() {
	args := argsT{}
	arg.MustParse(&args)

	scenarios, err := harness.LoadScenarios(args.Scenario)
	if err != nil {
		fmt.Printf("Could not load scenarios: %+v\n", err)
		os.Exit(1)
	}
	if err := harness.RunScenarios(scenarios); err != nil {
		fmt.Printf("Error while running scenarios: %+v\n", err)
		os.Exit(1)
	}
}
