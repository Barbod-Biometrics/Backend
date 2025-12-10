package exception

const (
	ErrorCodeInternal           ErrorCode = "INTERNAL_ERROR"
	ErrorCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrorCodeExternalService    ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)
