package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"

	"glow.dev.maio.me/seanj/vault-init/test/harness"
)

func main() {
	args := argsT{}
	arg.MustParse(&args)

	scenarios, err := harness.LoadScenarios(args.Scenario)
	if err != nil {
		fmt.Printf("Could not load scenarios: %+v\n", err)
		os.Exit(1)
	}
	harness.RunScenarios(scenarios)
}
