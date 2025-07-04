// Package strutl fornece funções utilitárias para manipulação de strings.
//
// Este pacote implementa diversas funções para transformação, formatação e
// manipulação de strings seguindo as melhores práticas de desenvolvimento Go.
package strutl

import (
	"sync"
	"unicode/utf8"
)

// Definição de variáveis globais
var (
	// uppercaseAcronym armazena os acrônimos configurados pelo usuário
	uppercaseAcronym = sync.Map{}
)

// Len retorna o número de runas em uma string.
// É um alias para utf8.RuneCountInString que lida corretamente com caracteres UTF-8.
func Len(s string) int {
	return utf8.RuneCountInString(s)
}

// ConfigureAcronym permite adicionar palavras que serão consideradas acrônimos.
// Útil para customizar o comportamento de funções como ToCamel e ToLowerCamel.
func ConfigureAcronym(key, val string) {
	uppercaseAcronym.Store(key, val)
}

// GetAcronym recupera um acrônimo previamente configurado.
// Retorna o valor do acrônimo e um booleano indicando se ele existe.
func GetAcronym(key string) (string, bool) {
	val, ok := uppercaseAcronym.Load(key)
	if !ok {
		return "", false
	}
	return val.(string), true
}
