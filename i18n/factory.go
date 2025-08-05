package i18n

import (
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/i18n/cache"
	"golang.org/x/text/language"
)

// ProviderType representa o tipo de provider de internacionalização
type ProviderType string

const (
	// ProviderTypeBasic é o provider básico sem cache
	ProviderTypeBasic ProviderType = "basic"
	// ProviderTypeCached é o provider com cache
	ProviderTypeCached ProviderType = "cached"
)

// ProviderConfig contém as configurações para criar um provider
type ProviderConfig struct {
	// Type é o tipo de provider a ser criado
	Type ProviderType
	// TranslationsPath é o caminho para os arquivos de tradução
	TranslationsPath string
	// TranslationsFormat é o formato dos arquivos de tradução (json, yaml, etc)
	TranslationsFormat string
	// Languages é a lista de idiomas suportados
	Languages []string
	// CacheTTL é o tempo de vida do cache (apenas para ProviderTypeCached)
	CacheTTL int
	// CacheMaxSize é o tamanho máximo do cache (apenas para ProviderTypeCached)
	CacheMaxSize int
}

// Provider é a interface que todos os providers devem implementar
type Provider interface {
	// Translate retorna a tradução para uma chave
	Translate(key string, data map[string]interface{}) (string, error)
	// TranslatePlural retorna a tradução plural para uma chave
	TranslatePlural(key string, count interface{}, data map[string]interface{}) (string, error)
	// LoadTranslations carrega as traduções de um diretório
	LoadTranslations(path string, format string) error
	// GetLanguages retorna os idiomas configurados
	GetLanguages() []language.Tag
	// SetLanguages configura os idiomas suportados
	SetLanguages(languages ...string) error
}

// NewProvider cria uma nova instância de um provider baseado na configuração
func NewProvider(config ProviderConfig) (Provider, error) {
	basicProvider := NewBasicProvider()

	// Configura os idiomas
	if err := basicProvider.SetLanguages(config.Languages...); err != nil {
		return nil, fmt.Errorf("failed to set languages: %w", err)
	}

	// Carrega as traduções
	if err := basicProvider.LoadTranslations(config.TranslationsPath, config.TranslationsFormat); err != nil {
		return nil, fmt.Errorf("failed to load translations: %w", err)
	}

	// Se for provider básico, retorna diretamente
	if config.Type == ProviderTypeBasic {
		return basicProvider, nil
	}

	// Se for provider com cache, envolve o provider básico com cache
	if config.Type == ProviderTypeCached {
		return cache.NewCachedProvider(basicProvider, time.Duration(config.CacheTTL)*time.Second, config.CacheMaxSize), nil
	}

	return nil, fmt.Errorf("unknown provider type: %s", config.Type)
}
