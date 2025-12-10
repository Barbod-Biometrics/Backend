package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/localization"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

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

func formatPlaceholders(msg string, params ...interface{}) string {
	if len(params) == 0 || msg == "" {
		return msg
	}
	out := msg
	for i, p := range params {
		ph := "{" + strconv.Itoa(i) + "}"
		out = strings.ReplaceAll(out, ph, fmt.Sprint(p))
	}
	return out
}

func ErrorTranslationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		var lastErr error = c.Errors.Last().Err

		if appErr, ok := exception.GetApplicationError(lastErr); ok {
			// start a short span to trace localization attempts
			tracer := otel.Tracer("gateway/middleware/localization")
			ctx := c.Request.Context()
			ctx, span := tracer.Start(ctx, "localization.translate")
			defer span.End()

			lang := getLangFromRequest(c)
			span.SetAttributes(
				attribute.String("request.lang", lang),
				attribute.String("error.code", string(appErr.Code)),
				attribute.String("http.route", c.FullPath()),
				attribute.Int("http.status_code", appErr.StatusCode),
			)

			translated := localization.GetErrorMessage(lang, appErr.Code)
			if translated == "" {
				span.SetAttributes(attribute.Bool("localization.translated", false))
				span.AddEvent("translation.missing")
				translated = appErr.Message
			} else {
				span.SetAttributes(attribute.Bool("localization.translated", true))
				span.AddEvent("translation.found")
			}

			resp := gin.H{
				"error":  translated,
				"code":   appErr.Code,
				"detail": appErr.Details,
			}

			c.AbortWithStatusJSON(appErr.StatusCode, resp)
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "an internal error occurred",
		})
	}
}
