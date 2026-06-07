package cli

// UsageError marks an operator error (invalid flags, missing configuration).
// Such errors result in exit code 2.
type UsageError struct {
	Err error
}

func (e *UsageError) Error() string { return e.Err.Error() }
func (e *UsageError) Unwrap() error { return e.Err }
