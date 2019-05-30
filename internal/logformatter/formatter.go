package logformatter

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	// FormatDefault creates the default logrus.TextFormatter configuration
	FormatDefault = "default"

	// FormatPlain creates a barebones formatter with few features enabled
	FormatPlain = "plain"

	// FormatJSON creates a logrus.JSONFormatter
	FormatJSON = "json"
)

// Configure returns a configured logrus.Formatter based on
// the provided formatter name
func Configure(useFormatter string) (log.Formatter, error) {
	var formatter log.Formatter

	switch useFormatter {
	case FormatDefault:
		formatter = new(log.TextFormatter)
	case FormatPlain:
		formatter = &log.TextFormatter{
			DisableColors:          true,
			DisableLevelTruncation: false,
			DisableSorting:         true,
			ForceColors:            false,
		}
	case FormatJSON:
		formatter = &log.JSONFormatter{
			DisableTimestamp: false,
		}
	default:
		return nil, errors.Errorf("unknown formatter configuration: %s", formatter)
	}

	return formatter, nil
}
