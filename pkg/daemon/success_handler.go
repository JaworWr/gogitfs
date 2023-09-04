package daemon

// SuccessHandler is used to handle daemon initialization success.
type SuccessHandler interface {
	// HandleSuccess is called after the daemon has initialized sucessfully.
	HandleSuccess()
}
