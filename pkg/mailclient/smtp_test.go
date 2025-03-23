package mailclient_test

import (
	"testing"

	"github.com/alvii147/nymphadora-api/pkg/mailclient"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
)

func TestNewSMTPMailClient(t *testing.T) {
	t.Parallel()

	hostname := testkit.MustGenerateRandomString(12, true, true, true)
	port := 587
	username := testkit.GenerateFakeEmail()
	password := testkit.GenerateFakePassword()
	timeProvider := timekeeper.NewFrozenProvider()

	mailclient.NewSMTPClient(hostname, port, username, password, timeProvider)
}
