package errutils_test

import (
	"errors"
	"testing"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/stretchr/testify/require"
)

func FormatErrorCaller(err error, msgs ...any) error {
	return errutils.FormatError(err, msgs...)
}

func FormatErrorCallerWrapper(err error, msgs ...any) error {
	return errutils.FormatError(FormatErrorCaller(err, msgs...), msgs...)
}

func FormatErrorfCallerWrapper(err error, format string, args ...any) error {
	return errutils.FormatErrorf(FormatErrorfCaller(err, format, args...), format, args...)
}

func FormatErrorfCaller(err error, format string, args ...any) error {
	return errutils.FormatErrorf(err, format, args...)
}

func TestTrimFuncName(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name                string
		funcName            string
		wantTrimmedFuncName string
	}{
		{
			name:                "Function",
			funcName:            "github.com/username/project/dir/package.Func",
			wantTrimmedFuncName: "package.Func",
		},
		{
			name:                "Receiver",
			funcName:            "github.com/username/project/dir/package.Struct.Func",
			wantTrimmedFuncName: "package.Struct.Func",
		},
		{
			name:                "Pointer receiver",
			funcName:            "github.com/username/project/dir/package.(*Struct).Func",
			wantTrimmedFuncName: "package.(*Struct).Func",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			trimmedFuncName := errutils.TrimFuncName(testcase.funcName)
			require.Equal(t, testcase.wantTrimmedFuncName, trimmedFuncName)
		})
	}
}

func TestJoin(t *testing.T) {
	t.Parallel()

	joined := errutils.SprintJoin([]any{"deadbeef", 42}, ":")
	require.Equal(t, "deadbeef:42", joined)
}

func TestFormatError(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name          string
		err           error
		msgs          []any
		wantErrString string
	}{
		{
			name: "No messages",
			err:  errors.New("wrapped error"),
			msgs: nil,
			wantErrString: "errutils_test.FormatErrorCallerWrapper " +
				"-> errutils_test.FormatErrorCaller -> wrapped error",
		},
		{
			name: "Including messages",
			err:  errors.New("wrapped error"),
			msgs: []any{"included", "messages"},
			wantErrString: "errutils_test.FormatErrorCallerWrapper: " +
				"included messages -> " +
				"errutils_test.FormatErrorCaller: " +
				"included messages -> wrapped error",
		},
		{
			name: "Nil error",
			err:  nil,
			msgs: []any{"included", "messages"},
			wantErrString: "errutils_test.FormatErrorCallerWrapper: " +
				"included messages -> " +
				"errutils_test.FormatErrorCaller: included messages",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			err := FormatErrorCallerWrapper(testcase.err, testcase.msgs...)
			require.Equal(t, testcase.wantErrString, err.Error())
			if testcase.err != nil {
				require.ErrorIs(t, err, testcase.err)
			}
		})
	}
}

func TestFormatErrorf(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name          string
		err           error
		format        string
		args          []any
		wantErrString string
	}{
		{
			name:   "No formatted messages",
			err:    errors.New("wrapped error"),
			format: "",
			args:   nil,
			wantErrString: "errutils_test.FormatErrorfCallerWrapper " +
				"-> errutils_test.FormatErrorfCaller -> wrapped error",
		},
		{
			name:   "Including formatted messages",
			err:    errors.New("wrapped error"),
			format: "%s-%d",
			args:   []any{"deadbeef", 42},
			wantErrString: "errutils_test.FormatErrorfCallerWrapper: " +
				"deadbeef-42 -> errutils_test.FormatErrorfCaller: " +
				"deadbeef-42 -> wrapped error",
		},
		{
			name:   "Nil error",
			err:    nil,
			format: "%s-%d",
			args:   []any{"deadbeef", 42},
			wantErrString: "errutils_test.FormatErrorfCallerWrapper: " +
				"deadbeef-42 -> errutils_test.FormatErrorfCaller: " +
				"deadbeef-42",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			err := FormatErrorfCallerWrapper(testcase.err, testcase.format, testcase.args...)
			require.Equal(t, testcase.wantErrString, err.Error())
			if testcase.err != nil {
				require.ErrorIs(t, err, testcase.err)
			}
		})
	}
}
