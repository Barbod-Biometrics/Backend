package v1

import (
	"strings"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/localization"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func HandleValidationError(c *gin.Context, err error) bool {
	lang := getLangFromRequest(c)

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		details := translateValidationErrors(lang, validationErrs)
		c.JSON(400, gin.H{
			"error":   localization.GetCommonError(lang, "invalid_request_body"),
			"details": details,
		})
		return true
	}

	return false
}

func translateValidationErrors(lang string, errs validator.ValidationErrors) string {
	var messages []string

	for _, fieldError := range errs {
		fieldName := fieldError.Field()
		tag := fieldError.Tag()

		translatedField := localization.GetField(lang, "fields."+camelToSnake(fieldName))
		if translatedField == "" {
			translatedField = fieldName
		}

		msg := translatedField + ": " + tag
		messages = append(messages, msg)
	}

	return strings.Join(messages, "\n")
}

func camelToSnake(input string) string {
	var result strings.Builder
	for i, r := range input {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32) 
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func getLangFromRequest(c *gin.Context) string {
	if q := c.Query("lang"); q != "" {
		return q
	}
	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		return "en"
	}
	parts := strings.Split(lang, ",")
	return strings.TrimSpace(parts[0])
}
