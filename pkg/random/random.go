package random

import (
	"crypto/rand"
	"math/big"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
)

const (
	// CharsLowerAlpha includes all lowercase alphabets.
	CharsLowerAlpha = "abcdefghijklmnopqrstuvwxyz"
	// CharsUpperAlpha includes all uppercase alphabets.
	CharsUpperAlpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// CharsUpperAlpha includes all numeric characters.
	CharsNumeric = "0123456789"
)

// GenerateInt64 generates a random integer from the interval [0, max).
func GenerateInt64(maxInt int64) (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(maxInt))
	if err != nil {
		return 0, errutils.FormatError(err, "rand.Int failed")
	}

	return n.Int64(), nil
}

// GenerateString generates random string of given length.
// allowLowerAlpha, allowUpperAlpha, and allowNumeric can be used to include/exclude
// lowercase, uppercase, and numeric characters respectively.
func GenerateString(
	n int,
	allowLowerAlpha bool,
	allowUpperAlpha bool,
	allowNumeric bool,
) (string, error) {
	lowerAlpha := ""
	if allowLowerAlpha {
		lowerAlpha = CharsLowerAlpha
	}

	upperAlpha := ""
	if allowUpperAlpha {
		upperAlpha = CharsUpperAlpha
	}

	numeric := ""
	if allowNumeric {
		numeric = CharsNumeric
	}

	allowed := []rune(lowerAlpha + upperAlpha + numeric)
	randRunes := make([]rune, n)

	for i := range randRunes {
		n, err := GenerateInt64(int64(len(allowed)))
		if err != nil {
			return "", errutils.FormatError(err)
		}

		randRunes[i] = allowed[n]
	}

	randString := string(randRunes)

	return randString, nil
}
