package vaultclient

import (
	vaultApi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

func newSecret(client *Client, path string, apiSecret *vaultApi.Secret) *secret {
	return &secret{
		client: client,
		Path:   path,
		Secret: apiSecret,
	}
}

func (s *secret) Fetch() (*vaultApi.Secret, error) {
	logical := s.client.vaultClient.Logical()
	secret, err := logical.Read(s.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get secret at path: %s", s.Path)
	}

	return secret, nil
}

func (s *secret) HasNewerVersion() (bool, error) {
	currentVersion, ok := s.Data["metadata"].(map[string]interface{})["version"].(int64)
	if !ok {
		return false, errors.Errorf("could not get current metadata for secret: data: %#v", s.Data)
	}

	next, err := s.Fetch()
	if err != nil {
		return false, errors.Wrap(err, "could not fetch secret from Vault")
	}

	nextVersion, ok := next.Data["metadata"].(map[string]interface{})["version"].(int64)
	if !ok {
		return false, errors.Errorf("could not get new metadata for secret: data: %#v", s.Data)
	}

	return currentVersion < nextVersion, nil
}
