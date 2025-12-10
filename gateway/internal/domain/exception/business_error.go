package exception

type BusinessError struct {
	*ApplicationError
}

func NewBusinessError(code ErrorCode, message string, err error) *BusinessError {
	return &BusinessError{ApplicationError: NewApplicationError(code, message, err)}
}

func (b *BusinessError) Unwrap() error {
	if b == nil {
		return nil
	}
	return b.ApplicationError
}
