package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"glow.dev.maio.me/seanj/vault-init/initializer"
)

var log = logrus.WithField("stream", "main")

func main() {
	config := initializer.NewConfig()
	if err := config.ValidateAndSetDefaults(); err != nil {
		log.WithError(err).Fatalf("Error validating configuration")
		os.Exit(1)
	}

	// Set log level according to verbosity
	if *config.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Set trace if Debug is set
	if *config.Debug {
		logrus.SetLevel(logrus.TraceLevel)
	}

	// Check if command is set
	if config.Command == nil {
		log.Fatalf("Command is required but not provided")
		os.Exit(1)
	}

	initializer.Run(context.Background(), config)
}
