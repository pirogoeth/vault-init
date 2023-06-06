package watcher

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/pirogoeth/vault-init/pkg/secret"
	"github.com/pirogoeth/vault-init/pkg/template"
	"github.com/pirogoeth/vault-init/pkg/vaultclient"
)

// Watcher watches the Client's secrets for updates and sends updated values
// to the supervisor.
type Watcher struct {
	client          vaultclient.VaultClient
	refreshDuration time.Duration
}

func NewWatcher(client vaultclient.VaultClient, refreshDuration time.Duration) (*Watcher, error) {
	return &Watcher{
		client:          client,
		refreshDuration: refreshDuration,
	}, nil
}

// Watch watches the secrets held in Client, sending updates through the update channel
func (w *Watcher) Watch(ctx context.Context, updateCh chan []string) {
	log.Infof("Watching secrets for updates every %s", w.refreshDuration.String())

	// Do the initial fetch and send it as an initial update to
	// the supervisor
	secrets, err := w.client.FetchSecrets()
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
func (w *Watcher) checkSecrets(secrets []*secret.Secret) (bool, error) {
	log.Debugf("Checking secret versions")

	updated := false

	for _, sec := range secrets {
		// Skip renewable secrets
		renewable, err := sec.IsRenewable()
		if renewable {
			log.WithField("secretPath", sec.Path).Debugf("Skipping secret as it is renewable")
			continue
		} else if err != nil {
			return false, errors.Wrapf(err, "could not check if secret `%s` is renewable", sec.Path)
		}

		nextSecret, err := w.client.FetchSecret(sec.Path)
		if err != nil {
			log.WithField("secretPath", sec.Path).WithError(err).Errorf("Error fetching secret for update check")
			return false, errors.Wrapf(err, "could not fetch secret `%s` for update check", sec.Path)
		}

		didUpdate, err := sec.Update(nextSecret)
		if err != nil {
			log.WithField("secretPath", sec.Path).WithError(err).Errorf("Error checking secret for updates")
			return false, errors.Wrapf(err, "could not check secret `%s` for updates", sec.Path)
		}

		if didUpdate {
			log.WithField("secretPath", sec.Path).Debugf("Update found for secrets")
			updated = true
		}
	}

	return updated, nil
}

// sendSecrets serializes all known secrets into environment templates
// and sends them as an update to the supervisor
func (w *Watcher) sendSecrets(updateCh chan []string, secrets []*secret.Secret) error {
	dataMap, err := secret.SecretsAsMap(secrets)
	if err != nil {
		return errors.Wrap(err, "could not convert secrets into data map")
	}

	dataMap, err = w.client.InjectChildContext(dataMap)
	if err != nil {
		return errors.Wrap(err, "could not inject child context from client")
	}

	environ, err := template.RenderEnvironmentFromDataMap(w.client.GetConfig(), dataMap)
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
