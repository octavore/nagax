package httperror

import "errors"

func UnwrapAll(err error) error {
	// note: does not work with Join'ed errors, eg fmt
	unwrappedErr := err
	for errors.Unwrap(unwrappedErr) != nil {
		unwrappedErr = errors.Unwrap(unwrappedErr)
	}
	return unwrappedErr
}
