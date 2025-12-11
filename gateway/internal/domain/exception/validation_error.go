package exception

import (
	"bytes"
	"strings"
)

type ValidationErrors struct {
	Errors []FieldError
}

func (ve ValidationErrors) Error() string {
	buff := bytes.NewBufferString("")
	for i := 0; i < len(ve.Errors); i++ {
		buff.WriteString(ve.Errors[i].Error())
		buff.WriteString("\n")
	}
	return strings.TrimSpace(buff.String())
}

func (v *ValidationErrors) Add(field, tag string) {
	v.Errors = append(v.Errors, FieldError{Field: field, Tag: tag})
}
