package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/alexflint/go-arg"
	"github.com/sirupsen/logrus"

	"glow.dev.maio.me/seanj/vault-init/internal/logformatter"
	"glow.dev.maio.me/seanj/vault-init/internal/supervise"
	"glow.dev.maio.me/seanj/vault-init/internal/vaultclient"
)

var log = logrus.WithField("stream", "main")

func main() {
	args := &args{}
	arg.MustParse(args)

	// Perform some checks on args and set defaults where applicable
	if err := args.CheckAndSetDefaults(); err != nil {
		log.Fatalf("Error while checking args: %v", err)
		os.Exit(1)
	}

	// Set log level according to verbosity
	if *args.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Set trace if Debug is set
	if *args.Debug {
		logrus.SetLevel(logrus.TraceLevel)
	}

	formatter, err := logformatter.Configure(args.LogFormat)
	if err != nil {
		log.WithError(err).Fatalf("Error configuring log formatter")
		os.Exit(1)
	}

	logrus.SetFormatter(formatter)

	// Make a context for controlling goroutines
	ctx, cancel := context.WithCancel(context.Background())

	// Load vaultclient-specific args into the vaultclient.Config struct
	vaultCfg := vaultclient.NewConfigWithDefaults()
	vaultCfg.AccessPolicies = args.AccessPolicies
	vaultCfg.DisableTokenRenew = *args.DisableTokenRenew
	vaultCfg.NoInheritToken = *args.NoInheritToken
	vaultCfg.OrphanToken = *args.OrphanToken
	vaultCfg.Paths = args.Paths
	vaultCfg.TokenPeriod = args.TokenPeriod
	vaultCfg.TokenTTL = args.TokenTTL

	// Read common Vault client configuration variables from environment,
	// storing them into the embedded `vaultApi.Config`
	if err := vaultCfg.ReadEnvironment(); err != nil {
		log.WithError(err).Fatalf("Could not create config from environment")
	}

	// Initialize the vaultclient wrapper
	vaultClient, err := vaultclient.NewClient(vaultCfg)
	if err != nil {
		log.WithError(err).Fatalf("Could not initialize vaultclient")
	}

	// Perform an initial health check on the Vault client
	if err := vaultClient.Check(); err != nil {
		log.WithError(err).Fatalf("Could not communicate with Vault")
	}

	// Create the child token and downgrade the Vault client to use it
	tokenDisplayName := fmt.Sprintf("Generated by vault-init for process: %s", args.Command)
	childSecret, err := vaultClient.CreateChildToken(tokenDisplayName)
	if err != nil {
		log.WithError(err).Fatalf("Could not create child token")
	}

	// Extract the secret and accessor
	accessor, err := childSecret.TokenAccessor()
	if err != nil {
		log.WithError(err).Fatalf("Could not get child token's accessor ID")
	}

	secret, err := childSecret.TokenID()
	if err != nil {
		log.WithError(err).Fatalf("Could not get child token's secret ID")
	}

	// Downgrade the Vault client to use the child token
	log.WithField("accessor_id", accessor).Infof("Downgrading Vault client to use child token")
	if err := vaultClient.SetToken(secret); err != nil {
		log.WithError(err).Fatalf("Could not use child token")
	}

	// Overwrite the VAULT_TOKEN environment variable with the child token
	// to prevent leaking the parent to the child process
	os.Setenv("VAULT_TOKEN", secret)

	// Start the child token renewer
	if err := childSecret.StartRenewer(); err != nil {
		log.WithError(err).Fatalf("Could not start child token renewer")
	}

	defer childSecret.StopRenewer()

	// Start the secret watcher
	log.WithField("paths", args.Paths).Debugf("Starting secrets watcher")
	updateCh, err := vaultClient.StartWatcher(ctx, *args.RefreshDuration)
	if err != nil {
		log.WithError(err).Fatalf("Could not start secrets watcher")
	}

	defer close(updateCh)

	// Configure the process supervisor
	supervisorCfg := &supervise.Config{
		Command:       args.Command,
		DisableReaper: *args.NoReaper,
		OneShot:       *args.OneShot,
	}

	// Create the supervisor with the configuration
	supervisor := supervise.NewSupervisor(supervisorCfg)

	go waitForSignal(ctx, cancel)

	// Launch the supervisor
	if err := supervisor.Start(ctx, updateCh); err != nil {
		log.WithError(err).Errorf("Supervisor returned an error")
	}

	// Cleanup and shutdown
	log.Infof("vault-init shutting down")

	// Revoke the child token
	err = childSecret.Revoke()
	if err != nil {
		log.WithError(err).Errorf("Could not revoke child token")
	} else {
		log.WithField("accessor_id", accessor).Debugf("Child token has been revoked")
	}
}

func waitForSignal(ctx context.Context, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	for {
		select {
		case sig := <-sigChan:
			log.Infof("Received signal %s, stopping", sig)
			cancel()
			return
		}
	}
}
