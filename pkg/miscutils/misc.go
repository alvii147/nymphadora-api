package miscutils

import "strings"

// EmailMaskingLength is the number of asterisks to use when masking email addresses.
const EmailMaskingLength = 5

// MaskEmail masks a given email by replacing most of the username with asterisks (*).
func MaskEmail(email string) string {
	mask := strings.Repeat("*", EmailMaskingLength)
	splitN := 2

	emailSplit := strings.SplitN(email, "@", splitN)
	if len(emailSplit) != splitN {
		return mask
	}

	username, remaining := emailSplit[0], emailSplit[1]

	var usernameStart string
	var usernameEnd string

	if len(username) > 0 {
		usernameStart = string(username[0])
		usernameEnd = string(username[len(username)-1])
	}

	return usernameStart + mask + usernameEnd + "@" + remaining
}
