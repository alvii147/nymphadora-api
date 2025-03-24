package miscutils_test

import (
	"testing"

	"github.com/alvii147/nymphadora-api/pkg/miscutils"
	"github.com/stretchr/testify/require"
)

func TestMaskEmail(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		email     string
		wantEmail string
	}{
		"Regular email": {
			email:     "davos.seaworth@westeros.com",
			wantEmail: "d*****h@westeros.com",
		},
		"Email with 2-character username": {
			email:     "ds@westeros.com",
			wantEmail: "d*****s@westeros.com",
		},
		"Email with 1-character username": {
			email:     "d@westeros.com",
			wantEmail: "d*****d@westeros.com",
		},
		"Email with no username": {
			email:     "@westeros.com",
			wantEmail: "*****@westeros.com",
		},
		"Invalid email": {
			email:     "davos.seaworthATwesteros.com",
			wantEmail: "*****",
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			maskedEmail := miscutils.MaskEmail(testcase.email)
			require.Equal(t, testcase.wantEmail, maskedEmail)
		})
	}
}
