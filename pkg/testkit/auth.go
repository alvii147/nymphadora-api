package testkit

import (
	"fmt"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/random"
)

// MustGenerateRandomString generates a random string and panics on error.
func MustGenerateRandomString(
	length int,
	allowLowerAlpha bool,
	allowUpperAlpha bool,
	allowNumeric bool,
) string {
	s, err := random.GenerateString(length, allowLowerAlpha, allowUpperAlpha, allowNumeric)
	if err != nil {
		panic(errutils.FormatError(err))
	}

	return s
}

// GenerateFakeEmail generates a randomized email addresss.
func GenerateFakeEmail() string {
	return fmt.Sprintf(
		"%s@%s.%s",
		MustGenerateRandomString(12, true, false, true),
		MustGenerateRandomString(10, true, false, false),
		MustGenerateRandomString(3, true, false, false),
	)
}

// GenerateFakeEmail generates a randomized password.
func GenerateFakePassword() string {
	password := MustGenerateRandomString(20, true, true, true)

	return password
}
