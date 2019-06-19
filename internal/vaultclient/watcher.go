package vaultclient

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
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

	w.sendSecrets(updateCh, secrets)

	for {
		select {
		case <-ctx.Done():
			log.Infof("Secret watcher exiting")
			return
		case <-time.After(w.refreshDuration):
			updated, err := w.checkSecrets(secrets)
			if err != nil {
				log.WithError(err).Errorf("Could not check secrets")
			}

			if updated {
				err := w.sendSecrets(updateCh, secrets)
				if err != nil {
					log.WithError(err).Errorf("Could not send secrets update")
				}

				log.Debugf("Successfully sent secrets update to supervisor")
			}
		}
	}
}

// checkSecrets iterates over all known secrets. _Only_ non-renewable
// secrets need to be monitored.
func (w *watcher) checkSecrets(secrets []*secret) (bool, error) {
	log.Debugf("Checking secret versions")

	updated := false

	for _, secret := range secrets {
		// Skip renewable secrets
		renewable, err := secret.IsRenewable()
		if renewable {
			log.WithField("secretPath", secret.Path).Debugf("Skipping secret as it is renewable")
			continue
		} else if err != nil {
			return false, errors.Wrapf(err, "could not check if secret `%s` is renewable", secret.Path)
		}

		hasUpdate, err := secret.Update()
		if err != nil {
			log.WithField("secretPath", secret.Path).WithError(err).Errorf("Error checking secret for updates")
			return false, errors.Wrapf(err, "could not check secret `%s` for updates", secret.Path)
		}

		if hasUpdate {
			log.WithField("secretPath", secret.Path).Debugf("Update found for secrets")
			updated = true
		}
	}

	return updated, nil
}

// sendSecrets serializes all known secrets into environment templates
// and sends them as an update to the supervisor
func (w *watcher) sendSecrets(updateCh chan []string, secrets []*secret) error {
	environ, err := w.client.secretsIntoEnviron(secrets)
	if err != nil {
		return errors.Wrap(err, "could not convert secrets into environment map")
	}

	vars := make([]string, 0)
	for key, value := range environ {
		vars = append(vars, strings.Join([]string{key, value}, "="))
	}

	updateCh <- vars

	return nil
}
