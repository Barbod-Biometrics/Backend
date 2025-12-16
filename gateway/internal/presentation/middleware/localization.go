package middleware

import (
	"net/http"
	"strings"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/localization"
	"github.com/gin-gonic/gin"
)

const TranslatorContextKey = "translator"

type LocalizationMiddleware struct {
	translator localization.Translator
}

func NewLocalization(translator localization.Translator) *LocalizationMiddleware {
	return &LocalizationMiddleware{
		translator: translator,
	}
}

func (lm *LocalizationMiddleware) Localization() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := getLocale(c.Request)

		translatorInstance := lm.translator.GetTranslator(locale)
		c.Set(TranslatorContextKey, translatorInstance)

		c.Next()
	}
}

func getLocale(request *http.Request) string {
	if q := request.URL.Query().Get("lang"); q != "" {
		return normalizeLocale(q)
	}

	header := request.Header.Get("Accept-Language")
	if header == "" {
		return "en_US"
	}
	parts := strings.Split(header, ",")
	if len(parts) > 0 {
		return normalizeLocale(strings.TrimSpace(parts[0]))
	}
	return "en_US"
}

func normalizeLocale(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	switch s {
	case "", "fa", "fa_ir", "fa-ir":
		return "fa_IR"
	case "en", "en_us", "en-us":
		return "en_US"
	default:
		if strings.HasPrefix(s, "fa") {
			return "fa_IR"
		}
		if strings.HasPrefix(s, "en") {
			return "en_US"
		}
		return "en_US"
	}
}

func GetTranslator(c *gin.Context) localization.TranslatorInstance {
	if translator, exists := c.Get(TranslatorContextKey); exists {
		if t, ok := translator.(localization.TranslatorInstance); ok {
			return t
		}
	}
	return nil
}
