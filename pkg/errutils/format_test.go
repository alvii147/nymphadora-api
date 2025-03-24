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

	testcases := map[string]struct {
		funcName            string
		wantTrimmedFuncName string
	}{
		"Function": {
			funcName:            "github.com/username/project/dir/package.Func",
			wantTrimmedFuncName: "package.Func",
		},
		"Receiver": {
			funcName:            "github.com/username/project/dir/package.Struct.Func",
			wantTrimmedFuncName: "package.Struct.Func",
		},
		"Pointer receiver": {
			funcName:            "github.com/username/project/dir/package.(*Struct).Func",
			wantTrimmedFuncName: "package.(*Struct).Func",
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
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

	testcases := map[string]struct {
		err           error
		msgs          []any
		wantErrString string
	}{
		"No messages": {
			err:  errors.New("wrapped error"),
			msgs: nil,
			wantErrString: "errutils_test.FormatErrorCallerWrapper " +
				"-> errutils_test.FormatErrorCaller -> wrapped error",
		},
		"Including messages": {
			err:  errors.New("wrapped error"),
			msgs: []any{"included", "messages"},
			wantErrString: "errutils_test.FormatErrorCallerWrapper: " +
				"included messages -> " +
				"errutils_test.FormatErrorCaller: " +
				"included messages -> wrapped error",
		},
		"Nil error": {
			err:  nil,
			msgs: []any{"included", "messages"},
			wantErrString: "errutils_test.FormatErrorCallerWrapper: " +
				"included messages -> " +
				"errutils_test.FormatErrorCaller: included messages",
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
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

	testcases := map[string]struct {
		err           error
		format        string
		args          []any
		wantErrString string
	}{
		"No formatted messages": {
			err:    errors.New("wrapped error"),
			format: "",
			args:   nil,
			wantErrString: "errutils_test.FormatErrorfCallerWrapper " +
				"-> errutils_test.FormatErrorfCaller -> wrapped error",
		},
		"Including formatted messages": {
			err:    errors.New("wrapped error"),
			format: "%s-%d",
			args:   []any{"deadbeef", 42},
			wantErrString: "errutils_test.FormatErrorfCallerWrapper: " +
				"deadbeef-42 -> errutils_test.FormatErrorfCaller: " +
				"deadbeef-42 -> wrapped error",
		},
		"Nil error": {
			err:    nil,
			format: "%s-%d",
			args:   []any{"deadbeef", 42},
			wantErrString: "errutils_test.FormatErrorfCallerWrapper: " +
				"deadbeef-42 -> errutils_test.FormatErrorfCaller: " +
				"deadbeef-42",
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := FormatErrorfCallerWrapper(testcase.err, testcase.format, testcase.args...)
			require.Equal(t, testcase.wantErrString, err.Error())
			if testcase.err != nil {
				require.ErrorIs(t, err, testcase.err)
			}
		})
	}
}
