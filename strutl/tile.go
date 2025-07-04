package strutl

import "strings"

// Tile repete o padrão até que o resultado atinja o tamanho 'length'.
// Retorna uma string vazia se o padrão é vazio ou o comprimento é <= 0.
//
// Exemplos:
//
//	Tile("ab", 5) -> "abababa" (truncado para "ababa")
//	Tile("abc", 8) -> "abcabcab"
//	Tile("-", 3) -> "---"
//	Tile("", 5) -> ""
//	Tile("x", 0) -> ""
func Tile(pattern string, length int) string {
	patLen := Len(pattern)
	if len(pattern) == 0 || length <= 0 {
		return ""
	}

	// Otimização para caso especial: se o padrão é um único caractere
	if patLen == 1 && len(pattern) == 1 {
		return strings.Repeat(pattern, length)
	}

	// Calcula quantas vezes o padrão completo cabe no comprimento desejado
	repeatCount := length / patLen
	remainder := length % patLen

	// Cria o resultado repetindo o padrão e adicionando parte do padrão se necessário
	var buff strings.Builder
	buff.Grow(length)

	for i := 0; i < repeatCount; i++ {
		buff.WriteString(pattern)
	}

	// Adiciona a parte restante do padrão se necessário
	if remainder > 0 {
		buff.WriteString(pattern[:remainder])
	}

	return buff.String()
}
