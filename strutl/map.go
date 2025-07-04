package strutl

import (
	"strings"
)

// MapLines executa a função fn em cada linha da string.
// Divide a string pelo caractere de nova linha ("\n"), executa 'fn' para cada linha
// e retorna a nova string combinando essas linhas com "\n".
//
// Exemplos:
//
//	MapLines("linha1\nlinha2", strings.ToUpper) -> "LINHA1\nLINHA2"
func MapLines(str string, fn func(string) string) string {
	return SplitAndMap(str, "\n", fn)
}

// SplitAndMap divide a string e executa a função fn em cada parte.
//
// Exemplos:
//
//	SplitAndMap("a,b,c", ",", strings.ToUpper) -> "A,B,C"
//	SplitAndMap("tag1|tag2|tag3", "|", func(s string) string { return "prefix-" + s }) -> "prefix-tag1|prefix-tag2|prefix-tag3"
func SplitAndMap(str, split string, fn func(string) string) string {
	if str == "" {
		return str
	}

	// Otimização: se não houver o separador na string, apenas aplica a função à string inteira
	if !strings.Contains(str, split) {
		return fn(str)
	}

	arr := strings.Split(str, split)
	for i := range arr {
		arr[i] = fn(arr[i])
	}

	return strings.Join(arr, split)
}
