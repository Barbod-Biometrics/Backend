package exception

import "fmt"

const (
	fieldErrMsg = "Error:Field validation for '%s' failed on the '%s' tag"
)

type FieldError struct {
	Field string
	Tag   string
}

func (fe FieldError) Error() string {
	return fmt.Sprintf(fieldErrMsg, fe.Field, fe.Tag)
}
