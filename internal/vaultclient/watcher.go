package vaultclient

import (
	"context"
	"strings"
	"time"

	vaultApi "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

func newWatcher(client *Client, refreshDuration time.Duration) (*watcher, error) {
	return &watcher{
		client:          client,
		refreshDuration: refreshDuration,
	}, nil
}

func (w *watcher) Watch(ctx context.Context, updateCh chan []string) {
	log.Infof("Watching secrets for updates every %s", w.refreshDuration.String())

	// Do the initial fetch and send it as an initial update to
	// the supervisor
	secrets, err := w.client.fetchSecrets()
	if err != nil {
		log.WithError(err).Fatalf("Could not collect secrets while starting watcher")
	}

	environ, err := w.client.secretsIntoEnviron(secrets)
	if err != nil {
		log.WithError(err).Fatalf("Could not convert secrets into environment map")
	}

	vars := make([]string, 0)
	for key, value := range environ {
		vars = append(vars, strings.Join([]string{key, value}, "="))
	}

	updateCh <- vars
	w.secrets = secrets

	for {
		select {
		case <-ctx.Done():
			log.Infof("Secret watcher exiting")
			return
		case <-time.After(w.refreshDuration):
			w.checkSecrets(ctx)
		}
	}
}

// checkSecrets iterates over all known secrets
func (w *watcher) checkSecrets(ctx context.Context) ([]*vaultApi.Secret, error) {
	log.Debugf("Checking secret versions")
	return nil, nil
}
