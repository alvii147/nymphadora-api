package errutils

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// TrimFuncName takes a full function name, including the package names,
// and trims up until the most immediate package and function name.
// For example, "github.com/username/project/dir/package.Func"
// becomes "package.Func".
func TrimFuncName(funcName string) string {
	_, trimmedFuncName := filepath.Split(funcName)

	return trimmedFuncName
}

// getFuncName retrieves full function name of the caller of the function
// that calls getFuncName.
func getFuncName() string {
	skipLevel := 2
	pc, _, _, _ := runtime.Caller(skipLevel)
	frames := runtime.CallersFrames([]uintptr{pc})
	frame, _ := frames.Next()
	funcName := frame.Function

	return funcName
}

// SprintJoin concatenates generic elements using a separator.
func SprintJoin(elems []any, sep string) string {
	strElems := make([]string, len(elems))
	for i, elem := range elems {
		strElems[i] = fmt.Sprint(elem)
	}

	joined := strings.Join(strElems, sep)

	return joined
}

// FormatError returns an error wrapped with the function name and messages.
//
//nolint:wrapcheck
func FormatError(err error, msgs ...any) error {
	funcName := getFuncName()
	trimmedFuncName := TrimFuncName(funcName)
	joinedMsgs := SprintJoin(msgs, " ")
	if len(joinedMsgs) > 0 {
		joinedMsgs = ": " + joinedMsgs
	}

	if err != nil {
		return fmt.Errorf("%s%s -> %w", trimmedFuncName, joinedMsgs, err)
	}

	return errors.New(trimmedFuncName + joinedMsgs)
}

// FormatErrorf returns an error wrapped with the function name and formatted messages.
//
//nolint:wrapcheck
func FormatErrorf(err error, format string, args ...any) error {
	funcName := getFuncName()
	trimmedFuncName := TrimFuncName(funcName)
	msg := fmt.Sprintf(format, args...)
	if len(msg) > 0 {
		msg = ": " + msg
	}

	if err != nil {
		return fmt.Errorf("%s%s -> %w", trimmedFuncName, msg, err)
	}

	return errors.New(trimmedFuncName + msg)
}
