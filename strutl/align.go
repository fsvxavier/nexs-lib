package strutl

import (
	"strings"
)

// AlignType define o tipo de alinhamento de texto como centro, esquerda ou direita.
type AlignType string

// Constantes para os tipos de alinhamento disponíveis para uso com a função Align.
const (
	Center AlignType = "center" // Alinha o texto ao centro
	Left   AlignType = "left"   // Alinha o texto à esquerda
	Right  AlignType = "right"  // Alinha o texto à direita
)

// Align alinha uma string de acordo com o tipo de alinhamento especificado.
// Deve ser um dos seguintes:
//   - strutl.Center: Centraliza o texto
//   - strutl.Left: Alinha o texto à esquerda
//   - strutl.Right: Alinha o texto à direita
//
// Exemplos:
//
//	Align("hello", Center, 10) -> "  hello   "
//	Align("hello", Left, 10) -> "hello"
//	Align("hello", Right, 10) -> "     hello"
func Align(str string, alignTo AlignType, width int) string {
	switch alignTo {
	case Center:
		return AlignCenter(str, width)
	case Right:
		return AlignRight(str, width)
	case Left:
		return AlignLeft(str)
	default:
		return str
	}
}

// AlignLeft alinha a string à esquerda. Para isso, remove espaços à esquerda de cada linha.
//
// Exemplos:
//
//	AlignLeft("  hello  ") -> "hello  "
//	AlignLeft("  linha1\n    linha2") -> "linha1\nlinha2"
func AlignLeft(str string) string {
	return MapLines(str, func(line string) string {
		return strings.TrimLeft(line, " ")
	})
}

// AlignRight alinha a string à direita. Remove espaços e adiciona preenchimento à esquerda
// de todas as linhas do texto com espaços até o tamanho de width.
//
// Exemplos:
//
//	AlignRight("hello", 10) -> "     hello"
//	AlignRight("linha1\nlinha2", 10) -> "     linha1\n     linha2"
func AlignRight(str string, width int) string {
	return MapLines(str, func(line string) string {
		line = strings.Trim(line, " ")
		return PadLeft(line, width, " ")
	})
}

// AlignCenter centraliza a string. Remove espaços e depois centraliza todas as linhas
// do texto com espaços.
//
// Exemplos:
//
//	AlignCenter("hello", 10) -> "  hello   "
//	AlignCenter("linha1\nlinha2", 10) -> "  linha1  \n  linha2  "
func AlignCenter(str string, width int) string {
	return MapLines(str, func(line string) string {
		line = strings.Trim(line, " ")
		return CenterText(line, width)
	})
}

// CenterText centraliza o texto adicionando espaços à esquerda e à direita.
// Assume que o texto é uma única linha. Para múltiplas linhas, use AlignCenter.
//
// Exemplos:
//
//	CenterText("hello", 11) -> "   hello   "
//	CenterText("center", 10) -> "  center  "
func CenterText(str string, width int) string {
	return Pad(str, width, " ", " ")
}
