package exception

type ResourceError struct {
	*ApplicationError
}

func NewResourceError(code ErrorCode, message string, err error) *ResourceError {
	return &ResourceError{ApplicationError: NewApplicationError(code, message, err)}
}

func (r *ResourceError) Unwrap() error {
	if r == nil {
		return nil
	}
	return r.ApplicationError
}
