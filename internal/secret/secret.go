package secret

import (
	"strings"

	"github.com/davecgh/go-spew/spew"
	vaultApi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

// Secret wraps the contents of a Secret from the Vault API
type Secret struct {
	*vaultApi.Secret

	// renewer is a reference to the running vaultApi.Renewer
	renewer *vaultApi.Renewer

	// Path is the logical path at which this secret was found
	Path string
}

// WrapChildToken wraps a special-case token that is injected into the child program.
func WrapChildToken(secret *vaultApi.Secret) *Secret {
	return New("CHILD_TOKEN", secret)
}

// New creates an instance of a Secret wrapped from the Vault API
func New(path string, secret *vaultApi.Secret) *Secret {
	log.Tracef(spew.Sprintf("creating new secret with data: %#v", secret.Data))
	return &Secret{
		Path:   path,
		Secret: secret,
	}
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

// Update receives a potentially new version of the inner secret, compares
// it to the version that is currently stored, and updates the stored secret
// if the metadata versions are different. Returns whether the secret was updated.
func (s *Secret) Update(nextSecret *Secret) (bool, error) {
	hasChanged, err := CompareSecretMetadata(s.Secret, nextSecret.Secret)
	if err != nil {
		return false, errors.Wrap(err, "could not compare current and next secret")
	}

	if hasChanged {
		s.Secret = nextSecret.Secret
	}

	return hasChanged, nil
}

// GetRenewer returns the associated renewer.
func (s *Secret) GetRenewer() *vaultApi.Renewer {
	return s.renewer
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

	return data
}

// WatchRenewer starts the renewer watcher for this secret.
func (s *Secret) WatchRenewer(renewer *vaultApi.Renewer) {
	log.Debugf("Watching renewer for secret `%s`", s.Path)
	s.renewer = renewer

	for {
		select {
		case err := <-renewer.DoneCh():
			if err != nil {
				log.WithError(err).WithField("secretPath", s.Path).Errorf("Renewer finished with an error")
			}

			log.WithField("secretPath", s.Path).Debugf("Renewer finished cleanly")
			s.renewer = nil
			return
		case output := <-renewer.RenewCh():
			s.Secret = output.Secret
			log.Debugf("Renewer updated secret `%s`", s.Path)
		}
	}
}
