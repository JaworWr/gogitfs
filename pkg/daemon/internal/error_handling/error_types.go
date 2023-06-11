package error_handling

type subprocessError struct {
	Msg string
}

func (e *subprocessError) Error() string {
	return e.Msg
}

type subprocessErrorWrapper struct {
	HasError bool
	Err      subprocessError
}

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

type UnknownError struct{}

func (_ *UnknownError) Error() string {
	return "unknown error in daemon - see log file"
}
