package validate_test

import (
	"testing"

	"github.com/alvii147/nymphadora-api/pkg/validate"
	"github.com/stretchr/testify/require"
)

func TestValidateStringNotBlank(t *testing.T) {
	t.Parallel()

	field := "value"

	testcases := map[string]struct {
		value      string
		wantPassed bool
	}{
		"Non-blank string": {
			value:      "d34d B33F",
			wantPassed: true,
		},
		"Empty string": {
			value:      "",
			wantPassed: false,
		},
		"Blank string": {
			value:      "   ",
			wantPassed: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := validate.NewValidator()
			v.ValidateStringNotBlank(field, testcase.value)
			require.Equal(t, testcase.wantPassed, v.Passed())

			failures := v.Failures()
			if testcase.wantPassed {
				require.Empty(t, failures)

				return
			}

			require.NotEmpty(t, failures[field])
		})
	}
}

func TestValidateStringMaxLength(t *testing.T) {
	t.Parallel()

	field := "value"

	testcases := map[string]struct {
		value      string
		maxLen     int
		wantPassed bool
	}{
		"String with allowed length": {
			value:      "d34d B33F",
			maxLen:     10,
			wantPassed: true,
		},
		"Empty string": {
			value:      "t00L0ng",
			maxLen:     5,
			wantPassed: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := validate.NewValidator()
			v.ValidateStringMaxLength(field, testcase.value, testcase.maxLen)
			require.Equal(t, testcase.wantPassed, v.Passed())

			failures := v.Failures()
			if testcase.wantPassed {
				require.Empty(t, failures)

				return
			}

			require.NotEmpty(t, failures[field])
		})
	}
}

func TestValidateStringMinLength(t *testing.T) {
	t.Parallel()

	field := "value"

	testcases := map[string]struct {
		value      string
		minLen     int
		wantPassed bool
	}{
		"String with allowed length": {
			value:      "d34d B33F",
			minLen:     5,
			wantPassed: true,
		},
		"Empty string": {
			value:      "t00sH0rt",
			minLen:     10,
			wantPassed: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := validate.NewValidator()
			v.ValidateStringMinLength(field, testcase.value, testcase.minLen)
			require.Equal(t, testcase.wantPassed, v.Passed())

			failures := v.Failures()
			if testcase.wantPassed {
				require.Empty(t, failures)

				return
			}

			require.NotEmpty(t, failures[field])
		})
	}
}

func TestValidateStringEmail(t *testing.T) {
	t.Parallel()

	field := "value"

	testcases := map[string]struct {
		value      string
		wantPassed bool
	}{
		"Valid email": {
			value:      "name@example.com",
			wantPassed: true,
		},
		"Invalid email": {
			value:      "1nv4l1d3m41l",
			wantPassed: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := validate.NewValidator()
			v.ValidateStringEmail(field, testcase.value)
			require.Equal(t, testcase.wantPassed, v.Passed())

			failures := v.Failures()
			if testcase.wantPassed {
				require.Empty(t, failures)

				return
			}

			require.NotEmpty(t, failures[field])
		})
	}
}

func TestValidateStringSlug(t *testing.T) {
	t.Parallel()

	field := "value"

	testcases := map[string]struct {
		value      string
		wantPassed bool
	}{
		"Valid slug": {
			value:      "d34d-b33f",
			wantPassed: true,
		},
		"String with invalid characters": {
			value:      "hello w*rld",
			wantPassed: false,
		},
		"String with beginning with hyphen": {
			value:      "-d34d-b33f",
			wantPassed: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := validate.NewValidator()
			v.ValidateStringSlug(field, testcase.value)
			require.Equal(t, testcase.wantPassed, v.Passed())

			failures := v.Failures()
			if testcase.wantPassed {
				require.Empty(t, failures)

				return
			}

			require.NotEmpty(t, failures[field])
		})
	}
}

func TestValidateStringOptions(t *testing.T) {
	t.Parallel()

	field := "value"

	testcases := map[string]struct {
		value         string
		options       []string
		caseSensitive bool
		wantPassed    bool
	}{
		"String in options, case sensitive": {
			value:         "deadbeef",
			options:       []string{"lorem", "deadbeef", "ipsum"},
			caseSensitive: true,
			wantPassed:    true,
		},
		"String not in options, case sensitive": {
			value:         "deadbeef",
			options:       []string{"lorem", "ipsum"},
			caseSensitive: true,
			wantPassed:    false,
		},
		"String in options but wrong case, case sensitive": {
			value:         "deadbeef",
			options:       []string{"lorem", "DeAdBeEf", "ipsum"},
			caseSensitive: true,
			wantPassed:    false,
		},
		"String in options, case insensitive": {
			value:         "deadbeef",
			options:       []string{"lorem", "deadbeef", "ipsum"},
			caseSensitive: false,
			wantPassed:    true,
		},
		"String not in options, case insensitive": {
			value:         "deadbeef",
			options:       []string{"lorem", "ipsum"},
			caseSensitive: false,
			wantPassed:    false,
		},
		"String in options but wrong case, case insensitive": {
			value:         "deadbeef",
			options:       []string{"lorem", "DeAdBeEf", "ipsum"},
			caseSensitive: false,
			wantPassed:    true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := validate.NewValidator()
			v.ValidateStringOptions(field, testcase.value, testcase.options, testcase.caseSensitive)
			require.Equal(t, testcase.wantPassed, v.Passed())

			failures := v.Failures()
			if testcase.wantPassed {
				require.Empty(t, failures)

				return
			}

			require.NotEmpty(t, failures[field])
		})
	}
}
