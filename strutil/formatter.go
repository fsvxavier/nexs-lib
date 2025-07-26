package strutil

import (
	"strings"
	"unicode"
)

// Alignment constants for text alignment operations
const (
	AlignLeft   = 0
	AlignCenter = 1
	AlignRight  = 2
)

// AlignType text align variable like center or left.
type AlignType string

// Align type constants to use with align function.
const (
	CenterAlign AlignType = "center"
	LeftAlign   AlignType = "left"
	RightAlign  AlignType = "right"
)

// Align aligns text with int alignment constants
func Align(text string, align int, width int) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		if len(line) >= width {
			result.WriteString(line)
			continue
		}

		switch align {
		case AlignLeft:
			result.WriteString(line + strings.Repeat(" ", width-len(line)))
		case AlignRight:
			result.WriteString(strings.Repeat(" ", width-len(line)) + line)
		case AlignCenter:
			totalPad := width - len(line)
			leftPad := totalPad / 2
			rightPad := totalPad - leftPad
			result.WriteString(strings.Repeat(" ", leftPad) + line + strings.Repeat(" ", rightPad))
		default:
			result.WriteString(line)
		}
	}

	return result.String()
}

// AlignText is a general method to align text to left/center/right
// each line in text using AlignType constants
func AlignText(text string, alignType AlignType, width int) string {
	switch alignType {
	case LeftAlign:
		return MapLines(text, func(line string) string {
			if len(line) >= width {
				return line
			}
			return line + strings.Repeat(" ", width-len(line))
		})
	case RightAlign:
		return MapLines(text, func(line string) string {
			if len(line) >= width {
				return line
			}
			return strings.Repeat(" ", width-len(line)) + line
		})
	case CenterAlign:
		return MapLines(text, func(line string) string {
			if len(line) >= width {
				return line
			}
			totalPad := width - len(line)
			leftPad := totalPad / 2
			rightPad := totalPad - leftPad
			return strings.Repeat(" ", leftPad) + line + strings.Repeat(" ", rightPad)
		})
	default:
		return text
	}
}

// Center aligns text to center
func Center(text string, width int) string {
	return Align(text, AlignCenter, width)
}

// AlignLeftText aligns string to the left. To achieve that it left trims every line.
func AlignLeftText(str string) string {
	return MapLines(str, func(line string) string {
		return strings.TrimLeft(line, " ")
	})
}

// AlignRightText aligns string to the right. It trims and left pads all the lines
// in the text with space to the size of width.
func AlignRightText(str string, width int) string {
	return MapLines(str, func(line string) string {
		line = strings.Trim(line, " ")
		return PadLeft(line, width, " ")
	})
}

// AlignCenterText centers str. It trims and then centers all the lines in the
// text with space.
func AlignCenterText(str string, width int) string {
	return MapLines(str, func(line string) string {
		line = strings.Trim(line, " ")
		return CenterText(line, width)
	})
}

// CenterText centers the text by adding spaces to the left and right.
// It assumes the text is one line. For multiple lines use AlignCenterText.
func CenterText(str string, width int) string {
	return Pad(str, width, " ", " ")
}

// PadLeft left pads a string str with "pad". The string is padded to
// the size of width.
func PadLeft(str string, width int, pad string) string {
	return Tile(pad, width-Len(str)) + str
}

// PadRight right pads a string str with "pad". The string is padded to
// the size of width.
func PadRight(str string, width int, pad string) string {
	return str + Tile(pad, width-Len(str))
}

// PadBoth pads both sides of the string with the specified padding string
func PadBoth(str string, pad string, length int) string {
	if len(str) >= length || len(pad) == 0 {
		return str
	}

	totalPad := length - len(str)
	rightPad := totalPad / 2
	leftPad := totalPad - rightPad // Left gets the extra char when odd

	return Tile(pad, leftPad) + str + Tile(pad, rightPad)
} // Pad left and right pads a string str with leftPad and rightPad. The string
// is padded to the size of width.
func Pad(str string, width int, leftPad, rightPad string) string {
	switch {
	case Len(leftPad) == 0:
		return PadRight(str, width, rightPad)
	case Len(rightPad) == 0:
		return PadLeft(str, width, leftPad)
	}
	padLen := (width - Len(str)) / 2
	return Tile(leftPad, padLen) + str + Tile(rightPad, width-Len(str)-padLen)
}

// Indent prefixes each line with the given indent string
func Indent(text string, indent string) string {
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")
	var result strings.Builder

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		// Only add indent to non-empty lines
		if line != "" {
			result.WriteString(indent + line)
		}
	}

	return result.String()
}

// ExpandTabs replaces tabs with spaces, properly calculating column positions
func ExpandTabs(text string, tabSize int) string {
	if tabSize <= 0 {
		return text
	}

	var result strings.Builder
	column := 0

	for _, r := range text {
		if r == '\t' {
			// Calculate how many spaces to the next tab stop
			spacesToAdd := tabSize - (column % tabSize)
			result.WriteString(strings.Repeat(" ", spacesToAdd))
			column += spacesToAdd
		} else if r == '\n' {
			result.WriteRune(r)
			column = 0 // Reset column on newline
		} else {
			result.WriteRune(r)
			column++
		}
	}

	return result.String()
}

// WordWrap wraps text to the specified width
func WordWrap(text string, width int) string {
	return WordWrapWithBreak(text, width, false)
}

// WordWrapWithBreak wraps text to the specified width with option to break long words
func WordWrapWithBreak(text string, width int, breakLongWords bool) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		// Handle long words that exceed width
		if breakLongWords && len(word) > width {
			// If we have content in current line, finish it first
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}

			// Break the long word into chunks
			for len(word) > width {
				lines = append(lines, word[:width])
				word = word[width:]
			}

			// Add remaining part of word if any
			if len(word) > 0 {
				currentLine.WriteString(word)
			}
			continue
		}

		if currentLine.Len() == 0 {
			currentLine.WriteString(word)
		} else if currentLine.Len()+1+len(word) <= width {
			currentLine.WriteString(" " + word)
		} else {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
} // DrawBox draws a box around the text
func DrawBox(text string, width int, height int) string {
	if width < 3 || height < 3 {
		return text
	}

	lines := strings.Split(text, "\n")
	boxWidth := width
	boxHeight := height

	var result strings.Builder

	// Top border
	result.WriteString("┌" + strings.Repeat("─", boxWidth-2) + "┐\n")

	// Content lines
	for i := 0; i < boxHeight-2; i++ {
		result.WriteString("│")

		if i < len(lines) {
			line := lines[i]
			if len(line) > boxWidth-2 {
				line = line[:boxWidth-2]
			}
			result.WriteString(line)
			result.WriteString(strings.Repeat(" ", boxWidth-2-len(line)))
		} else {
			result.WriteString(strings.Repeat(" ", boxWidth-2))
		}

		result.WriteString("│\n")
	}

	// Bottom border
	result.WriteString("└" + strings.Repeat("─", boxWidth-2) + "┘")

	return result.String()
}

// Box9Slice is used by DrawBox functions to draw frames around text content by
// defining the corner and edge characters.
type Box9Slice struct {
	Top         string
	TopRight    string
	Right       string
	BottomRight string
	Bottom      string
	BottomLeft  string
	Left        string
	TopLeft     string
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

// DefaultBox9Slice defines the character object to use with "CustomBox".
func DefaultBox9Slice() Box9Slice {
	return defaultBox9Slice
}

// SimpleBox9Slice defines a character set to use with DrawCustomBox using ASCII characters
func SimpleBox9Slice() Box9Slice {
	return simpleBox9Slice
}

// DrawCustomBox creates a frame with "content" in it using custom box characters
func DrawCustomBox(content string, width int, align AlignType, chars Box9Slice) string {
	if width < 3 {
		return content
	}

	lines := strings.Split(content, "\n")
	var result strings.Builder

	// Top border
	result.WriteString(chars.TopLeft)
	result.WriteString(strings.Repeat(chars.Top, width-2))
	result.WriteString(chars.TopRight + "\n")

	// Content lines
	for _, line := range lines {
		result.WriteString(chars.Left)

		// Apply alignment and padding
		var alignedLine string
		contentWidth := width - 2

		switch align {
		case LeftAlign:
			alignedLine = line
			if len(line) > contentWidth {
				alignedLine = line[:contentWidth]
			} else {
				alignedLine += strings.Repeat(" ", contentWidth-len(line))
			}
		case RightAlign:
			if len(line) > contentWidth {
				alignedLine = line[:contentWidth]
			} else {
				alignedLine = strings.Repeat(" ", contentWidth-len(line)) + line
			}
		case CenterAlign:
			if len(line) > contentWidth {
				alignedLine = line[:contentWidth]
			} else {
				totalPad := contentWidth - len(line)
				leftPad := totalPad / 2
				rightPad := totalPad - leftPad
				alignedLine = strings.Repeat(" ", leftPad) + line + strings.Repeat(" ", rightPad)
			}
		default:
			alignedLine = line
			if len(line) < contentWidth {
				alignedLine += strings.Repeat(" ", contentWidth-len(line))
			}
		}

		result.WriteString(alignedLine)
		result.WriteString(chars.Right + "\n")
	}

	// Bottom border
	result.WriteString(chars.BottomLeft)
	result.WriteString(strings.Repeat(chars.Bottom, width-2))
	result.WriteString(chars.BottomRight)

	return result.String()
}

// DrawBoxWithAlign draws a box around text with specified alignment
func DrawBoxWithAlign(content string, width int, align AlignType) string {
	return DrawCustomBox(content, width, align, defaultBox9Slice)
}

// DrawBoxSimple draws a simple box with the specified border character
func DrawBoxSimple(text string, border rune) string {
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	borderStr := string(border)
	totalWidth := maxLen + 4 // content + 2 spaces + 2 borders

	var result strings.Builder

	// Top border
	result.WriteString(strings.Repeat(borderStr, totalWidth) + "\n")

	// Content lines
	for _, line := range lines {
		result.WriteString(borderStr + " " + line)
		result.WriteString(strings.Repeat(" ", maxLen-len(line)) + " " + borderStr + "\n")
	}

	// Bottom border (without final newline)
	result.WriteString(strings.Repeat(borderStr, totalWidth))

	return result.String()
}

// Summary truncates text to maxlength and adds a suffix
func Summary(text string, maxlength int, suffix string) string {
	if maxlength <= 0 {
		return ""
	}

	if len(text) <= maxlength {
		return text
	}

	// If maxlength is too small for suffix, just return part of the text
	if maxlength < len(suffix) {
		return text[:maxlength]
	}

	// Try to break at word boundary
	targetLength := maxlength - len(suffix)
	truncated := text[:targetLength]

	// Look for last space to break at word boundary
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > 0 {
		// If we find a good word boundary, use it
		if lastSpace >= targetLength*2/3 { // Only use if reasonably close to target
			return strings.TrimRight(text[:lastSpace], " ") + suffix
		}
	}

	// No good word boundary, just cut at character level
	return text[:targetLength] + suffix
}

// Capitalize makes the first character uppercase
func Capitalize(text string) string {
	if text == "" {
		return ""
	}

	r := []rune(text)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
