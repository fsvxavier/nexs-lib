package hooks

// Observer define a interface para hooks do i18n
type Observer interface {
	// OnMissingTranslation é chamado quando uma tradução não é encontrada
	OnMissingTranslation(key string, language string)
	// OnTranslationLoaded é chamado quando um arquivo de traduções é carregado
	OnTranslationLoaded(language string, format string)
}

var observers []Observer

// RegisterHook registra um novo hook para eventos do i18n
func RegisterHook(observer Observer) {
	observers = append(observers, observer)
}

// NotifyMissingTranslation notifica todos os observers sobre uma tradução faltante
func NotifyMissingTranslation(key string, language string) {
	for _, observer := range observers {
		observer.OnMissingTranslation(key, language)
	}
}

// NotifyTranslationLoaded notifica todos os observers quando traduções são carregadas
func NotifyTranslationLoaded(language string, format string) {
	for _, observer := range observers {
		observer.OnTranslationLoaded(language, format)
	}
}
