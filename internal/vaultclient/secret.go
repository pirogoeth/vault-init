package vaultclient

import (
	"encoding/json"
	"strings"

	"github.com/davecgh/go-spew/spew"
	vaultApi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

func newSecret(client *Client, path string, apiSecret *vaultApi.Secret) *secret {
	log.Tracef(spew.Sprintf("creating new secret with data: %#v", apiSecret.Data))
	return &secret{
		client: client,
		Path:   path,
		Secret: apiSecret,
	}
}

// Fetch retrieves a new copy of the token, storing it in this secret
func (s *secret) Fetch() (*secret, error) {
	logical := s.client.vaultClient.Logical()
	secret, err := logical.Read(s.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get secret at path: %s", s.Path)
	}

	s.Secret = secret
	return s, nil
}

func (s *secret) Revoke() error {
	if s.Auth != nil {
		return s.revokeAuth()
	} else if s.LeaseID != "" {
		return s.revokeLease()
	}

	log.Warnf("Secret at `%s` can not be revoked; it is neither auth nor leased secret", s.Path)
	return nil
}

func (s *secret) revokeLease() error {
	if err := s.client.vaultClient.Sys().Revoke(s.LeaseID); err != nil {
		return errors.Wrapf(err, "could not revoke lease by id: %s", s.LeaseID)
	}

	return nil
}

func (s *secret) revokeAuth() error {
	tokenSys := s.client.vaultClient.Auth().Token()

	accessor, err := s.TokenAccessor()
	if err != nil {
		return errors.Wrapf(err, "could not get token accessor for secret loaded from path: %s", s.Path)
	}

	if err := tokenSys.RevokeAccessor(accessor); err != nil {
		return errors.Wrapf(err, "could not revoke token by accessor: %s: path %s", accessor, s.Path)
	}

	return nil
}

func (s *secret) IsRenewable() (bool, error) {
	var authRenewable bool
	var leaseRenewable bool

	leaseRenewable, err := s.TokenIsRenewable()
	if err != nil {
		return false, errors.Wrap(err, "error checking token renewability")
	}

	if s.Auth != nil {
		authRenewable = s.Auth.Renewable
	} else {
		authRenewable = false
	}

	return authRenewable || leaseRenewable, nil
}

// Update fetches the secret from the backend and compares it to the
// current secret. Returns whether or not the secret has changed.
func (s *secret) Update() (bool, error) {
	var shouldUpdate = false
	var err error

	metadata, ok := s.Data["metadata"].(map[string]interface{})
	if ok {
		shouldUpdate, err = s.metadataUpdate(metadata)
		if err != nil {
			return false, errors.Wrap(err, "could not update secret from metadata")
		}
	}

	return shouldUpdate, nil
}

func (s *secret) metadataUpdate(metadata map[string]interface{}) (bool, error) {
	currentVersionIface, ok := metadata["version"]
	if !ok {
		return false, errors.Errorf("could not get version from current secret's metadata")
	}

	currentVersionJSON := currentVersionIface.(json.Number)
	if !ok {
		return false, errors.Errorf("could not type assert metadata.version as json.Number")
	}

	currentVersion, err := currentVersionJSON.Int64()
	if err != nil {
		return false, errors.Errorf("could not convert metadata.version json.Number to int64")
	}

	next, err := s.Fetch()
	if err != nil {
		return false, errors.Wrap(err, "could not fetch secret from Vault")
	}

	nextMetadata, ok := next.Data["metadata"].(map[string]interface{})
	if !ok {
		return false, errors.Errorf("could not read metadata from new secret")
	}

	nextVersionIface, ok := nextMetadata["version"]
	if !ok {
		return false, errors.Errorf("could not get version from next secret's metadata")
	}

	nextVersionStr, ok := nextVersionIface.(json.Number)
	if !ok {
		return false, errors.Errorf("could not type assert nextMetadata.version as json.Number")
	}

	nextVersion, err := nextVersionStr.Int64()
	if err != nil {
		return false, errors.Errorf("could not convert nextMetadata.version json.Number to int")
	}

	return currentVersion < nextVersion, nil
}

func (s *secret) StartRenewer() error {
	renewable, err := s.IsRenewable()
	if err != nil {
		return errors.Wrap(err, "could not check if secret is renewable")
	}

	if !renewable {
		log.Debugf("Secret `%s` is not renewable; skipping renewer", s.Path)
		return nil
	}

	renewer, err := s.client.vaultClient.NewRenewer(&vaultApi.RenewerInput{
		Secret: s.Secret,
	})
	if err != nil {
		return errors.Wrap(err, "could not start secret renewer")
	}

	go renewer.Renew()
	go s.watchRenewer(renewer)
	s.renewer = renewer

	return nil
}

func (s *secret) StopRenewer() {
	s.renewer.Stop()
}

func (s *secret) dataMap() map[string]interface{} {
	data := s.Data
	pathComponents := strings.Split(s.Path, "/")
	for idx := range pathComponents {
		component := pathComponents[len(pathComponents)-1-idx]
		if component == "" {
			// Skip blank components
			continue
		}

		component = strings.ReplaceAll(component, "-", "_")

		tmp := make(map[string]interface{}, 0)
		tmp[component] = data
		data = tmp
	}

	log.Tracef(spew.Sprintf("nested secret data looks like: %#v", data))

	return data
}

func (s *secret) watchRenewer(renewer *vaultApi.Renewer) {
	log.Debugf("Watching renewer for secret `%s`", s.Path)

	for {
		select {
		case err := <-renewer.DoneCh():
			if err != nil {
				log.WithError(err).WithField("secretPath", s.Path).Errorf("Renewer finished with an error")
			}

			log.WithField("secretPath", s.Path).Debugf("Renewer finished cleanly")
			return
		case output := <-renewer.RenewCh():
			s.Secret = output.Secret
			log.Debugf("Renewer updated secret `%s`", s.Path)
		}
	}
}
