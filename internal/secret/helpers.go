package secret

import (
	"encoding/json"
	"reflect"

	vaultApi "github.com/hashicorp/vault/api"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

// HasMetadata determines if a secret has associated metadata.
func HasMetadata(sec *vaultApi.Secret) bool {
	val, ok := sec.Data["metadata"]
	if !ok {
		return false
	}

	return val != nil
}

// GetVersionFromSecretMetadata extracts the current version of the secret from associated metadata.
func GetVersionFromSecretMetadata(metadata map[string]interface{}) (int64, error) {
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

// CompareSecretMetadata takes two secrets and compares the version contained in the metadata.
func CompareSecretMetadata(current, next *vaultApi.Secret) (bool, error) {
	currentMeta := current.Data["metadata"].(map[string]interface{})
	nextMeta := next.Data["metadata"].(map[string]interface{})

	// Special case: some secrets do not have metadata (ie., otp secrets)
	if currentMeta == nil && nextMeta == nil {
		log.Errorf("current %+v next %+v do not have attached metadata", current, next)
		return false, nil
	}

	currentVersion, err := GetVersionFromSecretMetadata(currentMeta)
	if err != nil {
		return false, errors.Errorf("could not get version for current secret")
	}

	nextVersion, err := GetVersionFromSecretMetadata(nextMeta)
	if err != nil {
		return false, errors.Errorf("could not get version for next secret")
	}

	return currentVersion < nextVersion, nil
}

// CompareSecretData takes two secrets and compares them by their contents.
func CompareSecretData(current, next *vaultApi.Secret) (bool, error) {
	return !reflect.DeepEqual(current, next), nil
}

func SecretsAsMap(secrets []*Secret) (map[string]interface{}, error) {
	data := make(map[string]interface{}, 0)

	for _, secret := range secrets {
		if err := mergo.Merge(&data, secret.dataMap()); err != nil {
			return nil, errors.Wrap(err, "could not merge secret to data")
		}
	}

	return data, nil
}
