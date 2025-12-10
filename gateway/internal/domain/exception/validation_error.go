package exception

type ValidationError struct {
	*ApplicationError
}

func NewValidationError(message string, err error) *ValidationError {
	return &ValidationError{ApplicationError: NewApplicationError(ErrorCodeValidationFailed, message, err)}
}

func (v *ValidationError) Unwrap() error {
	if v == nil {
		return nil
	}
	return v.ApplicationError
}
