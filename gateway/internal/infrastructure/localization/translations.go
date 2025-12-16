package localization

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/localization"
	"github.com/go-playground/locales/en_US"
	"github.com/go-playground/locales/fa_IR"
	ut "github.com/go-playground/universal-translator"
)

var translationsMap = make(map[string]map[string]string)

var serviceInstance *TranslationService

func GetService() *TranslationService {
	if serviceInstance == nil {
		serviceInstance = NewTranslationService()
	}
	return serviceInstance
}

type TranslationService struct {
	mu sync.RWMutex
	uv *ut.UniversalTranslator
}

func NewTranslationService() *TranslationService {
	service := &TranslationService{
		uv: createUniversalTranslator(),
	}
	service.loadAndAddTranslations()
	return service
}

func createUniversalTranslator() *ut.UniversalTranslator {
	en := en_US.New()
	fa := fa_IR.New()
	return ut.New(en, en, fa)
}

func (t *TranslationService) loadAndAddTranslations() {
	addTranslations("fa_IR", Persian, t.uv)
	addTranslations("en_US", English, t.uv)
}

var _ localization.Translator = (*TranslationService)(nil)

func (t *TranslationService) GetTranslator(locale string) localization.TranslatorInstance {
	switch locale {
	case "", "fa":
		locale = "fa_IR"
	case "en":
		locale = "en_US"
	}

	translator, found := t.uv.GetTranslator(locale)
	if !found {
		translator, _ = t.uv.GetTranslator("en_US")
		locale = "en_US"
	}

	return &translatorInstance{
		translator: translator,
		locale:     locale,
	}
}

type translatorInstance struct {
	translator ut.Translator
	locale     string
}

func (ti *translatorInstance) Translate(key string, params ...string) (string, error) {
	translation, err := ti.translator.T(key)
	if err != nil {
		return key, err
	}

	for i, param := range params {
		placeholder := fmt.Sprintf("{%d}", i)
		translation = strings.ReplaceAll(translation, placeholder, param)
	}

	return translation, nil
}

func (ti *translatorInstance) Locale() string {
	return ti.locale
}

func AddToTranslationsMap(locale string, key string, value string) {
	if _, ok := translationsMap[locale]; !ok {
		translationsMap[locale] = make(map[string]string)
	}
	translationsMap[locale][key] = value
}

func addTranslations(locale string, translations map[string]interface{}, universalTranslator *ut.UniversalTranslator) {
	translator, found := universalTranslator.GetTranslator(locale)
	if !found {
		panic(fmt.Errorf("translator for locale %s not found", locale))
	}

	flattenedTranslations := loadTranslations(locale, translations)

	for key, translation := range flattenedTranslations {
		translator.Add(key, translation, true)
	}
}

func loadTranslations(locale string, translations map[string]interface{}) map[string]string {
	if translations, ok := translationsMap[locale]; ok {
		return translations
	}

	flattenedTranslations := make(map[string]string)
	flattenMap("", translations, flattenedTranslations)

	translationsMap[locale] = flattenedTranslations

	return flattenedTranslations
}

func flattenMap(prefix string, input map[string]interface{}, output map[string]string) {
	for k, v := range input {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}
		switch value := v.(type) {
		case map[string]interface{}:
			flattenMap(fullKey, value, output)
		case string:
			output[fullKey] = value
		default:

		}
	}
}
