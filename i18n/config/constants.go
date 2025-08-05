package config

// Supported translation file formats
const (
	FormatJSON = "json"
	FormatYAML = "yaml"
	FormatYML  = "yml"
)

// Default file extensions
const (
	ExtensionJSON = ".json"
	ExtensionYAML = ".yaml"
	ExtensionYML  = ".yml"
)

// Default directory structure
const (
	DefaultTranslationsDir = "translations"
	LCMessagesDir          = "LC_MESSAGES"
)

// Translation file name patterns
const (
	TranslationFilePattern = "translations.%s.%s" // e.g., translations.pt-BR.json
	GetTextFilePattern     = "%s.%s"              // e.g., messages.po
)
