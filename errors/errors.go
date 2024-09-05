package errors

import (
	stdErr "errors"
	"fmt"
	"runtime"
)

var RuntimeFileInfo = false

func As(err error, target any) bool {
	return stdErr.As(err, target)
}

func Is(err, target error) bool {
	return stdErr.Is(err, target)
}
func Join(errs ...error) error {
	return stdErr.Join(errs...)
}

func New(text string) error {
	return stdErr.New(text)
}

func Newf(text string, args ...any) error {
	return fmt.Errorf(text, args...)
}

func Unwrap(err error) error {
	return stdErr.Unwrap(err)
}

func Wrap(err error, msg string, args ...any) error {
	if err == nil {
		return err
	}
	if RuntimeFileInfo {
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			msg += " function=%s file=%s line=%d"
			rf := runtime.FuncForPC(pc)
			args = append(args, rf.Name(), file, line)
		}
	}

	msg += ": %w"
	args = append(args, err)

	return fmt.Errorf(msg, args...)
}
