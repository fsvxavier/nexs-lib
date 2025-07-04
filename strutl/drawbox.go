package strutl

import (
	"errors"
	"strings"
)

// Box9Slice é usado pelas funções DrawBox para desenhar quadros ao redor do conteúdo
// de texto, definindo os caracteres de canto e borda. Veja DefaultBox9Slice para
// um exemplo.
type Box9Slice struct {
	Top         string // Caractere da borda superior
	TopRight    string // Caractere do canto superior direito
	Right       string // Caractere da borda direita
	BottomRight string // Caractere do canto inferior direito
	Bottom      string // Caractere da borda inferior
	BottomLeft  string // Caractere do canto inferior esquerdo
	Left        string // Caractere da borda esquerda
	TopLeft     string // Caractere do canto superior esquerdo
}

var defaultBox9Slice = Box9Slice{
	Top:         "─",
	TopRight:    "┐",
	Right:       "│",
	BottomRight: "┘",
	Bottom:      "─",
	BottomLeft:  "└",
	Left:        "│",
	TopLeft:     "┌",
}

// DefaultBox9Slice define o objeto de caracteres a ser usado com "CustomBox".
// É usado como objeto Box9Slice na função "DrawBox".
//
// Exemplo:
//
//	DrawCustomBox("Hello World", 20, Center, DefaultBox9Slice(), "\n")
//
// Resultado:
//
//	┌──────────────────┐
//	│   Hello World    │
//	└──────────────────┘
func DefaultBox9Slice() Box9Slice {
	return defaultBox9Slice
}

var simpleBox9Slice = Box9Slice{
	Top:         "-",
	TopRight:    "+",
	Right:       "|",
	BottomRight: "+",
	Bottom:      "-",
	BottomLeft:  "+",
	Left:        "|",
	TopLeft:     "+",
}

// SimpleBox9Slice define um conjunto de caracteres a ser usado com DrawCustomBox.
// Utiliza apenas caracteres ASCII simples.
//
// Exemplo:
//
//	DrawCustomBox("Hello World", 20, Center, SimpleBox9Slice(), "\n")
//
// Resultado:
//
//	+------------------+
//	|   Hello World    |
//	+------------------+
func SimpleBox9Slice() Box9Slice {
	return simpleBox9Slice
}

// DrawCustomBox cria um quadro com "conteúdo" nele. Caracteres no quadro são especificados por "chars".
// "align" define o alinhamento do conteúdo. Deve ser uma das constantes de AlignType.
// Existem 2 objetos Box9Slice pré-definidos que podem ser recuperados por strutl.DefaultBox9Slice() ou
// strutl.SimpleBox9Slice()
//
// Exemplo:
//
//	DrawCustomBox("Hello World", 20, Center, SimpleBox9Slice(), "\n")
//
// Resultado:
//
//	+------------------+
//	|   Hello World    |
//	+------------------+
func DrawCustomBox(content string, width int, align AlignType, chars *Box9Slice, strNewLine string) (string, error) {
	nl := []byte("\n")
	if strNewLine != "" {
		nl = []byte(strNewLine)
	}

	topInsideWidth := width - Len(chars.TopLeft) - Len(chars.TopRight)
	middleInsideWidth := width - Len(chars.Left) - Len(chars.Right)
	bottomInsideWidth := width - Len(chars.BottomLeft) - Len(chars.BottomRight)
	if topInsideWidth < 1 || middleInsideWidth < 1 || bottomInsideWidth < 1 {
		return "", errors.New("largura insuficiente")
	}

	content = WordWrap(content, middleInsideWidth, true)
	lines := strings.Split(content, "\n")

	var buff strings.Builder
	minNumBytes := (width + 1) * (len(lines) + 2)
	buff.Grow(minNumBytes)

	// top
	buff.WriteString(chars.TopLeft)
	buff.WriteString(Tile(chars.Top, topInsideWidth))
	buff.WriteString(chars.TopRight)
	buff.Write(nl)

	// middle
	left := []byte(chars.Left)
	right := []byte(chars.Right)
	for _, line := range lines {
		line = Align(line, align, middleInsideWidth)
		if align == Left {
			line = PadRight(line, middleInsideWidth, " ")
		}
		if line == "" {
			line = strings.Repeat(" ", middleInsideWidth)
		}

		buff.Write(left)
		buff.WriteString(line)
		buff.Write(right)
		buff.Write(nl)
	}

	// bottom
	buff.WriteString(chars.BottomLeft)
	buff.WriteString(Tile(chars.Bottom, bottomInsideWidth))
	buff.WriteString(chars.BottomRight)

	return buff.String(), nil
}

// DrawBox cria um quadro com "conteúdo" nele. O objeto DefaultBox9Slice é usado para
// definir os caracteres no quadro. "align" define o alinhamento do conteúdo.
// Deve ser uma das constantes de AlignType.
//
// Exemplo:
//
//	DrawBox("Hello World", 20, Center)
//
// Resultado:
//
//	┌──────────────────┐
//	│   Hello World    │
//	└──────────────────┘
func DrawBox(content string, width int, align AlignType) (string, error) {
	return DrawCustomBox(content, width, align, &defaultBox9Slice, "\n")
}
