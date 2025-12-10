package localization

import (
	"fmt"
	"strings"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	infraLoc "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/localization"
)

func mapLangToLocale(langKey string) string {
	switch langKey {
	case "fa", "fa_IR":
		return "fa_IR"
	default:
		return "en_US"
	}
}

func snakeToLowerCamel(input string) string {
	parts := strings.Split(strings.ToLower(input), "_")
	if len(parts) == 0 {
		return input
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		p := parts[i]
		if p == "" {
			continue
		}
		out += strings.ToUpper(p[:1]) + p[1:]
	}
	return out
}

func GetErrorMessage(langKey string, code exception.ErrorCode) string {
	svc := infraLoc.GetService()
	locale := mapLangToLocale(langKey)
	translator, _ := svc.GetTranslator(locale)

	candidates := []string{
		string(code),
		strings.ToLower(string(code)),
		"errors." + string(code),
		"errors." + strings.ToLower(string(code)),
		"errors." + snakeToLowerCamel(string(code)),
	}

	for _, key := range candidates {
		res, _ := translator.T(key)
		if res != key && res != "" {
			return res
		}
	}
	return ""
}

func GetErrorMessageWithArgs(langKey string, code exception.ErrorCode, args ...interface{}) string {
	svc := infraLoc.GetService()
	locale := mapLangToLocale(langKey)
	translator, _ := svc.GetTranslator(locale)

	strArgs := make([]string, len(args))
	for i, a := range args {
		strArgs[i] = fmt.Sprint(a)
	}

	candidates := []string{
		string(code),
		strings.ToLower(string(code)),
		"errors." + string(code),
		"errors." + strings.ToLower(string(code)),
		"errors." + snakeToLowerCamel(string(code)),
	}

	for _, key := range candidates {
		res, _ := translator.T(key, strArgs...)
		if res != key && res != "" {
			return res
		}
	}
	return ""
}

func GetField(langKey string, key string) string {
	svc := infraLoc.GetService()
	locale := mapLangToLocale(langKey)
	translator, _ := svc.GetTranslator(locale)
	res, _ := translator.T(key)
	if res == key {
		return ""
	}
	return res
}

func GetCommonError(langKey string, key string) string {
	svc := infraLoc.GetService()
	locale := mapLangToLocale(langKey)
	translator, _ := svc.GetTranslator(locale)
	tkey := "errors." + key
	res, _ := translator.T(tkey)
	if res == tkey {
		return ""
	}
	return res
}

func GetSuccessMessage(langKey string, key string) string {
	svc := infraLoc.GetService()
	locale := mapLangToLocale(langKey)
	translator, _ := svc.GetTranslator(locale)
	tkey := "successMessage." + key
	res, _ := translator.T(tkey)
	if res == tkey {
		return ""
	}
	return res
}

func AddTranslation(langKey string, code exception.ErrorCode, message string) {
	svc := infraLoc.GetService()
	locale := mapLangToLocale(langKey)
	translator, found := svc.GetTranslator(locale)
	if !found {
		return
	}
	translator.Add(string(code), message, true)
	translator.Add("errors."+snakeToLowerCamel(string(code)), message, true)
	infraLoc.AddToTranslationsMap(locale, string(code), message)
	infraLoc.AddToTranslationsMap(locale, "errors."+snakeToLowerCamel(string(code)), message)
}
