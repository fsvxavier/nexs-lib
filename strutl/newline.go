package strutl

import "runtime"

// OSNewLine retorna o caractere de nova linha padrão do sistema operacional.
// É \r\n no Windows e \n em outros sistemas.
//
// Exemplo:
//
//	OSNewLine() -> "\r\n" no Windows
//	OSNewLine() -> "\n" em outros sistemas
func OSNewLine() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}
