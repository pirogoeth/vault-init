package stringlist

import (
	"strings"
)

func Contains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if strings.Contains(item, needle) {
			return true
		}
	}

	return false
}
