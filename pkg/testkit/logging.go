package testkit

import (
	"bytes"
	"regexp"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/logging"
)

// MustParseLogMessage parses a given log message and panics on error.
func MustParseLogMessage(msg string) (string, time.Time, string, string) {
	r := regexp.MustCompile(`^\s*\[([DIWE])\]\s+(\d{4}\/\d{2}\/\d{2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+(.+)\s*$`)
	matches := r.FindStringSubmatch(msg)

	if len(matches) != 5 {
		panic(errutils.FormatError(nil, "regexp.Regexp.FindStringSubmatch failed"))
	}

	logLevel := matches[1]
	logTime, err := time.ParseInLocation("2006/01/02 15:04:05", matches[2], time.UTC)
	if err != nil {
		panic(errutils.FormatErrorf(nil, "time.ParseInLocation failed to parse time %s", matches[2]))
	}

	logFile := matches[3]
	logMsg := matches[4]

	return logLevel, logTime, logFile, logMsg
}

// CreateInMemLogger creates a new in-memory logger.
func CreateInMemLogger() (*bytes.Buffer, *bytes.Buffer, logging.Logger) {
	var bufOut bytes.Buffer
	var bufErr bytes.Buffer
	logger := logging.NewLogger(&bufOut, &bufErr)

	return &bufOut, &bufErr, logger
}
