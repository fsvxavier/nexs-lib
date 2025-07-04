package strutl

import "unicode/utf8"

// Reverse inverte os caracteres de uma string.
// Esta função manipula corretamente strings UTF-8, preservando caracteres multibyte.
//
// Exemplos:
//
//	Reverse("abc") -> "cba"
//	Reverse("こんにちは") -> "はちにんこ"
//	Reverse("Hello, 世界") -> "界世 ,olleH"
func Reverse(s string) string {
	// Caso trivial para strings vazias
	if s == "" {
		return s
	}

	// Conta o número de runas para pré-alocar o slice
	n := utf8.RuneCountInString(s)

	// Se n == len(s), então a string contém apenas ASCII, podemos usar um método mais simples
	if n == len(s) {
		// Versão otimizada para ASCII
		result := make([]byte, n)
		for i := 0; i < n; i++ {
			result[n-1-i] = s[i]
		}
		return string(result)
	}

	// Para strings com caracteres UTF-8, precisamos usar decodificação de runas
	// Pré-alocamos um slice de runas para a performance
	runes := make([]rune, n)

	// Preenchemos o slice de runas de trás para frente
	for _, r := range s {
		n--
		runes[n] = r
	}

	return string(runes)
}
