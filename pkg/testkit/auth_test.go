package testkit_test

import (
	"net/mail"
	"testing"

	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestMustGenerateRandomStringSuccess(t *testing.T) {
	t.Parallel()

	randomString := testkit.MustGenerateRandomString(8, true, true, true)
	require.Len(t, randomString, 8)
}

func TestMustGenerateRandomStringPanic(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		testkit.MustGenerateRandomString(8, false, false, false)
	})
}

func TestGenerateFakeEmail(t *testing.T) {
	t.Parallel()

	for range 10 {
		email := testkit.GenerateFakeEmail()
		_, err := mail.ParseAddress(email)
		require.NoError(t, err)
	}
}

func TestGenerateFakePassword(t *testing.T) {
	t.Parallel()

	for range 10 {
		password := testkit.GenerateFakePassword()
		require.NotEmpty(t, password)
	}
}
