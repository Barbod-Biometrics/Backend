package exception

const (
	ErrorCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrorCodeInvalidToken     ErrorCode = "INVALID_TOKEN"
	ErrorCodeTokenExpired     ErrorCode = "TOKEN_EXPIRED"
	ErrorCodeMissingToken     ErrorCode = "MISSING_TOKEN"
	ErrorCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrorCodeInsufficientRole ErrorCode = "INSUFFICIENT_ROLE"
)
