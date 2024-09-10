package errors

import (
	stdErr "errors"
	"fmt"
	"runtime"
)

func addRuntimeInfo(s *string) {
	pc, file, line, ok := runtime.Caller(1)
	if ok && s != nil {
		f := runtime.FuncForPC(pc)
		*s += fmt.Sprintf(" [function='%s' file='%s' line=%d]", f.Name(), file, line)
	}
}

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
	addRuntimeInfo(&text)
	return stdErr.New(text)
}

func Newf(text string, args ...any) error {
	addRuntimeInfo(&text)
	return fmt.Errorf(text, args...)
}

func Unwrap(err error) error {
	return stdErr.Unwrap(err)
}

func Wrap(err error, msg string, args ...any) error {
	if err == nil {
		return err
	}
	addRuntimeInfo(&msg)
	msg += ": %w"
	args = append(args, err)

	return fmt.Errorf(msg, args...)
}
