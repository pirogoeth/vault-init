package helpers

import (
	"os"
	"testing"
)

func SkipIfNotHarnessed(t *testing.T) error {
	_, exists := os.LookupEnv("VAULT_INIT_UNDER_TEST")
	if !exists {
		t.Skipf("Not testing from vault-init-test-harness, skipping")
	}

	return nil
}
