package secret

import (
	"encoding/json"
	"strings"

	"github.com/davecgh/go-spew/spew"
	vaultApi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"

	"glow.dev.maio.me/seanj/vault-init/internal/vaultclient"
)

// Secret wraps the contents of a Secret from the Vault API
type Secret struct {
	*vaultApi.Secret

	// client is a reference to the vaultclient.Client
	client vaultclient.VaultClient

	// renewer is a reference to the running vaultApi.Renewer
	renewer *vaultApi.Renewer

	// Path is the logical path at which this secret was found
	Path string
}

// WrapChildToken wraps a special-case token that is injected into the child program.
func WrapChildToken(client vaultclient.VaultClient, secret *vaultApi.Secret) *Secret {
	return NewSecret(client, "CHILD_TOKEN", secret)
}

// NewSecret creates an instance of a Secret wrapped from the Vault API
func NewSecret(client vaultclient.VaultClient, path string, secret *vaultApi.Secret) *Secret {
	log.Tracef(spew.Sprintf("creating new secret with data: %#v", secret.Data))
	return &Secret{
		client: client,
		Path:   path,
		Secret: secret,
	}
}

// Fetch retrieves a new copy of the token, storing it in this secret
func (s *Secret) Fetch() (*Secret, error) {
	// xxx - refactoring
	// logical := s.client.vaultClient.Logical()
	// secret, err := logical.Read(s.Path)
	secret, err := s.client.ReadLogical(s.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get secret at path: %s", s.Path)
	}

	s.Secret = secret
	return s, nil
}

// Revoke revokes the underlying secret that was retrieved from Vault.
func (s *Secret) Revoke() error {
	if s.Auth != nil {
		return s.revokeAuth()
	} else if s.LeaseID != "" {
		return s.revokeLease()
	}

	log.Warnf("Secret at `%s` can not be revoked; it is neither auth nor leased secret", s.Path)
	return nil
}

func (s *Secret) revokeLease() error {
	// xxx - refactoring
	// if err := s.client.vaultClient.Sys().Revoke(s.LeaseID); err != nil {
	if err := s.client.RevokeLease(s.LeaseID); err != nil {
		return errors.Wrapf(err, "could not revoke lease by id: %s", s.LeaseID)
	}

	return nil
}

func (s *Secret) revokeAuth() error {
	accessor, err := s.TokenAccessor()
	if err != nil {
		return errors.Wrapf(err, "could not get token accessor for secret loaded from path: %s", s.Path)
	}

	// xxx - refactoring
	// tokenSys := s.client.vaultClient.Auth().Token()
	// if err := tokenSys.RevokeAccessor(accessor); err != nil {
	if err := s.client.RevokeTokenAccessor(accessor); err != nil {
		return errors.Wrapf(err, "could not revoke token by accessor: %s: path %s", accessor, s.Path)
	}

	return nil
}

// IsRenewable detemines if the secret is renewable.
func (s *Secret) IsRenewable() (bool, error) {
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
func (s *Secret) Update() (bool, error) {
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

func getVersionFromMetadata(metadata map[string]interface{}) (int64, error) {
	versionIface, ok := metadata["version"]
	if !ok {
		return 0, errors.Errorf("could not get version from secret metadata: %#v", metadata)
	}

	versionJSON, ok := versionIface.(json.Number)
	if !ok {
		return 0, errors.Errorf("could not type assert metadata.version as json.Number")
	}

	version, err := versionJSON.Int64()
	if err != nil {
		return 0, errors.Wrapf(err, "could not convert metadata.version json.Number to int64")
	}

	return version, nil
}

func (s *Secret) metadataUpdate(metadata map[string]interface{}) (bool, error) {
	currentVersion, err := getVersionFromMetadata(metadata)
	if err != nil {
		return false, errors.Errorf("could not convert metadata.version json.Number to int64")
	}

	next, err := s.Fetch()
	if err != nil {
		return false, errors.Wrap(err, "could not fetch secret from Vault")
	}

	nextVersion, err := getVersionFromMetadata(
		next.Data["metadata"].(map[string]interface{}),
	)
	if err != nil {
		return false, errors.Errorf("could not convert nextMetadata.version json.Number to int")
	}

	return currentVersion < nextVersion, nil
}

// StartRenewer starts a goroutine that renews the underlying secret.
func (s *Secret) StartRenewer() error {
	renewable, err := s.IsRenewable()
	if err != nil {
		return errors.Wrap(err, "could not check if secret is renewable")
	}

	if !renewable {
		log.Debugf("Secret `%s` is not renewable; skipping renewer", s.Path)
		return nil
	}

	// xxx - refactoring
	// renewer, err := s.client.vaultClient.NewRenewer(&vaultApi.RenewerInput{
	// 	Secret: s.Secret,
	// })
	renewer, err := s.client.NewLeaseRenewer(&vaultApi.RenewerInput{
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

// StopRenewer stops the renewer attached to this secret.
func (s *Secret) StopRenewer() {
	s.renewer.Stop()
}

func (s *Secret) dataMap() map[string]interface{} {
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

func (s *Secret) watchRenewer(renewer *vaultApi.Renewer) {
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
