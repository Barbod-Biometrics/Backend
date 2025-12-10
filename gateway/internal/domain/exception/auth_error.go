package exception

type AuthError struct {
	*ApplicationError
}

func NewAuthError(code ErrorCode, message string, err error) *AuthError {
	return &AuthError{ApplicationError: NewApplicationError(code, message, err)}
}

func (a *AuthError) Unwrap() error {
	if a == nil {
		return nil
	}
	return a.ApplicationError
}
