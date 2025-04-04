package logging_test

import (
	"strings"
	"testing"

	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Parallel()

	bufOut, bufErr, logger := testkit.CreateInMemLogger()

	debugMessage := "Debug message"
	infoMessage := "Info message"
	warnMessage := "Warn message"
	errorMessage := "Error message"

	timeProvider := timekeeper.NewFrozenProvider()
	logger.LogDebug(debugMessage)
	logger.LogInfo(infoMessage)
	logger.LogWarn(warnMessage)
	logger.LogError(errorMessage)

	stdoutMessages := strings.Split(strings.TrimSpace(bufOut.String()), "\n")
	stderrMessages := strings.Split(strings.TrimSpace(bufErr.String()), "\n")

	require.Len(t, stdoutMessages, 3)
	require.Len(t, stderrMessages, 1)

	testcases := map[string]struct {
		capturedMessage string
		wantMessage     string
		wantLevel       string
	}{
		"Debug": {
			capturedMessage: stdoutMessages[0],
			wantMessage:     debugMessage,
			wantLevel:       "D",
		},
		"Info": {
			capturedMessage: stdoutMessages[1],
			wantMessage:     infoMessage,
			wantLevel:       "I",
		},
		"Warning": {
			capturedMessage: stdoutMessages[2],
			wantMessage:     warnMessage,
			wantLevel:       "W",
		},
		"Error": {
			capturedMessage: stderrMessages[0],
			wantMessage:     errorMessage,
			wantLevel:       "E",
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logLevel, logTime, logFile, logMsg := testkit.MustParseLogMessage(testcase.capturedMessage)
			require.Equal(t, testcase.wantLevel, logLevel)
			require.WithinDuration(t, logTime, timeProvider.Now(), testkit.TimeToleranceTentative)
			require.Contains(t, logFile, "pkg/logging/logging_test.go")
			require.Equal(t, testcase.wantMessage, logMsg)
		})

		break
	}
}
