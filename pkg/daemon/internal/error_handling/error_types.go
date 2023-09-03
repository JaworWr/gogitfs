// Package error_handling deals with sending errors from the daemon to the parent process.
package error_handling

import "errors"

type subprocessError struct {
	Msg string
}

func (e *subprocessError) Error() string {
	return e.Msg
}

// subprocessErrorWrapper is the actual value sent from the subprocess.
type subprocessErrorWrapper struct {
	HasError bool
	Err      subprocessError
}

// wrapError checks if err == nil and sets the HasError flag accordingly in the result.
func wrapError(err error) *subprocessErrorWrapper {
	res := subprocessErrorWrapper{HasError: false}
	if err == nil {
		return &res
	}
	res.HasError = true
	res.Err = subprocessError{err.Error()}
	return &res
}

func (w *subprocessErrorWrapper) unwrap() error {
	if !w.HasError {
		return nil
	}
	return &w.Err
}

// UnknownError is a generic error for subprocess failures.
var UnknownError = errors.New("unknown error in daemon - see log file")
