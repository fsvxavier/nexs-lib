package strutl

// PadLeft adiciona preenchimento à esquerda de uma string str com "pad".
// A string é preenchida até o tamanho de width.
//
// Exemplos:
//
//	PadLeft("hello", 10, " ") -> "     hello"
//	PadLeft("123", 5, "0") -> "00123"
//	PadLeft("abc", 6, "xy") -> "xyxyabc"
func PadLeft(str string, width int, pad string) string {
	strLen := Len(str)
	if strLen >= width {
		return str
	}

	return Tile(pad, width-strLen) + str
}

// PadRight adiciona preenchimento à direita de uma string str com "pad".
// A string é preenchida até o tamanho de width.
//
// Exemplos:
//
//	PadRight("hello", 10, " ") -> "hello     "
//	PadRight("123", 5, "0") -> "12300"
//	PadRight("abc", 6, "xy") -> "abcxyx"
func PadRight(str string, width int, pad string) string {
	strLen := Len(str)
	if strLen >= width {
		return str
	}

	return str + Tile(pad, width-strLen)
}

// Pad adiciona preenchimento à esquerda e à direita de uma string str com
// leftPad e rightPad. A string é preenchida até o tamanho de width.
//
// Exemplos:
//
//	Pad("hello", 10, " ", " ") -> "  hello   "
//	Pad("center", 10, "-", "-") -> "--center--"
//	Pad("text", 8, "<", ">") -> "<<text>>"
func Pad(str string, width int, leftPad, rightPad string) string {
	strLen := Len(str)
	if strLen >= width {
		return str
	}

	switch {
	case Len(leftPad) == 0:
		return PadRight(str, width, rightPad)
	case Len(rightPad) == 0:
		return PadLeft(str, width, leftPad)
	}

	// Calcula a quantidade de preenchimento necessária para cada lado
	padLen := (width - strLen) / 2
	// Para o caso de width - strLen ser ímpar, adicionamos um caractere extra à direita
	rightPadLen := width - strLen - padLen

	return Tile(leftPad, padLen) + str + Tile(rightPad, rightPadLen)
}
