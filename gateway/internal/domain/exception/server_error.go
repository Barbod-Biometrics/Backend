package exception

type ServerError struct {
	*ApplicationError
}

func NewServerError(code ErrorCode, message string, err error) *ServerError {
	return &ServerError{ApplicationError: NewApplicationError(code, message, err)}
}

func (s *ServerError) Unwrap() error {
	if s == nil {
		return nil
	}
	return s.ApplicationError
}
