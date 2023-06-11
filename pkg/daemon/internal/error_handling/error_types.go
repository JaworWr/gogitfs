package error_handling

type subprocessError struct {
	msg string
}

func (e *subprocessError) Error() string {
	return e.msg
}

type subprocessErrorWrapper struct {
	hasError bool
	err      subprocessError
}

func wrapError(err error) *subprocessErrorWrapper {
	res := subprocessErrorWrapper{hasError: false}
	if err == nil {
		return &res
	}
	res.hasError = true
	res.err = subprocessError{err.Error()}
	return &res
}

func (w *subprocessErrorWrapper) unwrap() error {
	if !w.hasError {
		return nil
	}
	return &w.err
}

type UnknownError struct{}

func (_ *UnknownError) Error() string {
	return "unknown error in daemon - see log file"
}
