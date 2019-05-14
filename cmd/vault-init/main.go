package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/alexflint/go-arg"
	log "github.com/sirupsen/logrus"

	"glow.dev.maio.me/seanj/vault-init/internal/supervise"
	"glow.dev.maio.me/seanj/vault-init/internal/vaultclient"
)

func main() {
	args := &args{}
	arg.MustParse(args)

	// Perform some checks on args and set defaults where applicable
	if err := args.CheckAndSetDefaults(); err != nil {
		log.Fatalf("Error while checking args: %v", err)
		os.Exit(1)
	}

	// Set log level according to verbosity
	if *args.InitVerbose {
		log.SetLevel(log.DebugLevel)
	}

	// Make a context for controlling goroutines
	ctx, cancel := context.WithCancel(context.Background())

	// Load vaultclient-specific args into the vaultclient.Config struct
	vaultCfg := vaultclient.NewConfigWithDefaults()
	vaultCfg.AccessPolicies = args.InitAccessPolicies
	vaultCfg.DisableTokenRenew = *args.InitDisableTokenRenew
	vaultCfg.OrphanToken = *args.InitOrphanToken
	vaultCfg.Paths = args.InitPaths
	vaultCfg.TokenRenew = *args.InitTokenRenew
	vaultCfg.TokenTTL = args.InitTokenTTL

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
	childToken, err := vaultClient.CreateChildToken()
	if err != nil {
		log.WithError(err).Fatalf("Could not create child token")
	}

	// Extract the secret and accessor
	accessor, err := childToken.TokenAccessor()
	if err != nil {
		log.WithError(err).Fatalf("Could not get child token's accessor ID")
	}

	secret, err := childToken.TokenID()
	if err != nil {
		log.WithError(err).Fatalf("Could not get child token's secret ID")
	}

	// Downgrade the Vault client to use the child token
	log.WithField("accessor_id", accessor).Infof("Downgrading Vault client to use child token")
	if err := vaultClient.SetToken(secret); err != nil {
		log.WithError(err).Fatalf("Could not use child token")
	}

	// Start the child token renewer

	// Start the secret watcher
	log.WithField("paths", args.InitPaths).Debugf("Starting secrets watcher")
	updateCh, err := vaultClient.StartWatcher(ctx, *args.InitRefreshDuration)
	if err != nil {
		log.WithError(err).Fatalf("Could not start secrets watcher")
	}

	defer close(updateCh)

	// Configure the process supervisor
	supervisorCfg := &supervise.Config{
		Command:       args.Command,
		DisableReaper: *args.InitNoReaper,
	}

	// Create the supervisor with the configuration
	supervisor := supervise.NewSupervisor(supervisorCfg)

	// Launch the child process inside of the supervisor
	if err := supervisor.Start(ctx, updateCh); err != nil {
		log.WithError(err).Fatal("Supervisor loop failed")
	}

	waitForSignal(ctx, cancel)

	// Cleanup and shutdown
	log.Infof("Supervisor shutting down")
}

func waitForSignal(ctx context.Context, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case sig := <-sigChan:
			log.Infof("Received signal %s, stopping", sig)
			cancel()
			return
		case <-time.After(1 * time.Second):
			continue
		}
	}
}
