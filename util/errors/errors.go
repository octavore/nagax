package errors

import (
	"errors"
	"fmt"

	goerrors "github.com/go-errors/errors"
)

// New returns a new wrapped error with m as message
func New(m string, a ...any) error {
	return goerrors.Wrap(fmt.Errorf(m, a...), 1)
}

// Wrap an error e if it is not nil
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	var e *goerrors.Error
	if errors.As(err, &e) {
		return err
	}
	return goerrors.Wrap(err, 1)
}

// WrapS is like Wrap but skips 'skip' lines of trace
func WrapS(err error, skip int) error {
	if err == nil {
		return nil
	}
	var e *goerrors.Error
	if errors.As(err, &e) {
		return err
	}
	return goerrors.Wrap(err, skip+1)
}

func IsType(e1, e2 error) bool {
	return errors.Is(e1, e2) || errors.Is(e2, e1)
}

func TypeName(err error) string {
	var e *goerrors.Error
	if errors.As(err, &e) {
		return e.TypeName()
	}
	return fmt.Sprintf("%T", err)
}
