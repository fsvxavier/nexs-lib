package strutl

import "unicode/utf8"

// IsASCII verifica se todos os caracteres de uma string estão na tabela ASCII padrão.
// Um caractere ASCII tem seu valor entre 0 e 127.
//
// Exemplos:
//
//	IsASCII("hello") -> true
//	IsASCII("olá") -> false
//	IsASCII("123ABC") -> true
//	IsASCII("世界") -> false
func IsASCII(s string) bool {
	// setBits é usado para rastrear quais bits estão definidos nos bytes da string.
	// Um caractere é ASCII se todos os seus bytes têm o bit mais significativo (MSB) igual a 0.
	// Isso significa que todos os caracteres ASCII têm um valor menor que 128 (0x80).
	var setBits uint8

	for i := 0; i < len(s); i++ {
		setBits |= s[i]
	}

	// utf8.RuneSelf é igual a 128 (0x80), que é o primeiro valor não-ASCII
	return setBits < utf8.RuneSelf
}
