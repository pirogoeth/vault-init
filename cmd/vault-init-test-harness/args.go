package main

type argsT struct {
	// Scenario is a list of scenarios that are to be run
	Scenario []string `arg:"positional"`
}
