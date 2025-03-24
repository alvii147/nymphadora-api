package random_test

import (
	"testing"
	"unicode"

	"github.com/alvii147/nymphadora-api/pkg/random"
	"github.com/stretchr/testify/require"
)

func TestGenerateInt64(t *testing.T) {
	t.Parallel()

	n, err := random.GenerateInt64(42)
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, int64(0))
	require.Less(t, n, int64(42))
}

func getLowerUpperNumericCharCounts(s string) (int, int, int) {
	lowerAlphaCount := 0
	upperAlphaCount := 0
	numericCount := 0

	for _, c := range s {
		switch {
		case unicode.IsLower(c):
			lowerAlphaCount += 1
		case unicode.IsUpper(c):
			upperAlphaCount += 1
		case unicode.IsNumber(c):
			numericCount += 1
		default:
			continue
		}
	}

	return lowerAlphaCount, upperAlphaCount, numericCount
}

func TestGenerateString(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		n                 int
		allowLowerAlpha   bool
		allowUpperAlpha   bool
		allowNumericAlpha bool
	}{
		"16-letter string, allow all": {
			n:                 16,
			allowLowerAlpha:   true,
			allowUpperAlpha:   true,
			allowNumericAlpha: true,
		},
		"16-letter string, allow only uppercase and numeric": {
			n:                 16,
			allowLowerAlpha:   false,
			allowUpperAlpha:   true,
			allowNumericAlpha: true,
		},
		"16-letter string, allow only lowercase and numeric": {
			n:                 16,
			allowLowerAlpha:   true,
			allowUpperAlpha:   false,
			allowNumericAlpha: true,
		},
		"16-letter string, allow only alphabets": {
			n:                 16,
			allowLowerAlpha:   true,
			allowUpperAlpha:   true,
			allowNumericAlpha: false,
		},
		"16-letter string, allow only lowercase": {
			n:                 16,
			allowLowerAlpha:   true,
			allowUpperAlpha:   false,
			allowNumericAlpha: false,
		},
		"16-letter string, allow only uppercase": {
			n:                 16,
			allowLowerAlpha:   false,
			allowUpperAlpha:   true,
			allowNumericAlpha: false,
		},
		"16-letter string, allow only numeric": {
			n:                 16,
			allowLowerAlpha:   false,
			allowUpperAlpha:   false,
			allowNumericAlpha: true,
		},
		"256-letter string, allow all": {
			n:                 256,
			allowLowerAlpha:   true,
			allowUpperAlpha:   true,
			allowNumericAlpha: true,
		},
		"256-letter string, allow only uppercase and numeric": {
			n:                 256,
			allowLowerAlpha:   false,
			allowUpperAlpha:   true,
			allowNumericAlpha: true,
		},
		"256-letter string, allow only lowercase and numeric": {
			n:                 256,
			allowLowerAlpha:   true,
			allowUpperAlpha:   false,
			allowNumericAlpha: true,
		},
		"256-letter string, allow only alphabets": {
			n:                 256,
			allowLowerAlpha:   true,
			allowUpperAlpha:   true,
			allowNumericAlpha: false,
		},
		"256-letter string, allow only lowercase": {
			n:                 256,
			allowLowerAlpha:   true,
			allowUpperAlpha:   false,
			allowNumericAlpha: false,
		},
		"256-letter string, allow only uppercase": {
			n:                 256,
			allowLowerAlpha:   false,
			allowUpperAlpha:   true,
			allowNumericAlpha: false,
		},
		"256-letter string, allow only numeric": {
			n:                 256,
			allowLowerAlpha:   false,
			allowUpperAlpha:   false,
			allowNumericAlpha: true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s, err := random.GenerateString(
				testcase.n,
				testcase.allowLowerAlpha,
				testcase.allowUpperAlpha,
				testcase.allowNumericAlpha,
			)
			require.NoError(t, err)
			require.Len(t, s, testcase.n)

			lowerAlphaCount, upperAlphaCount, numericCount := getLowerUpperNumericCharCounts(s)
			if !testcase.allowLowerAlpha {
				require.Equal(t, 0, lowerAlphaCount)
			}

			if !testcase.allowUpperAlpha {
				require.Equal(t, 0, upperAlphaCount)
			}

			if !testcase.allowNumericAlpha {
				require.Equal(t, 0, numericCount)
			}
		})
	}
}
