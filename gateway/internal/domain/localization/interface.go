package localization

type Translator interface {
	GetTranslator(locale string) TranslatorInstance
}

type TranslatorInstance interface {
	Translate(key string, params ...string) (string, error)
	Locale() string
}
