package stringlist

import (
	"testing"
)

var (
	defaultHaystack = []string{"go", "test", "-v", "-json", "github.com/pirogoeth/vault-init/pkg/harness/util/stringlist"}
)

func TestContainsTrue(t *testing.T) {
	needle := "-json"

	if !Contains(defaultHaystack, needle) {
		t.Errorf(
			"haystack `%s` contains `%s` but Contains returned false",
			defaultHaystack,
			needle,
		)
	}
}

func BenchmarkContainsBest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Contains(defaultHaystack, defaultHaystack[0])
	}
}

func BenchmarkContainsWorst(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Contains(defaultHaystack, defaultHaystack[4])
	}
}
