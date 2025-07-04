package strutl

import (
	"strings"
	"unicode"
)

// toCamelInitCase converte uma string para CamelCase ou lowerCamelCase dependendo do parâmetro initCase.
// Se initCase for true, retorna CamelCase (primeira letra maiúscula).
// Se initCase for false, retorna lowerCamelCase (primeira letra minúscula).
func toCamelInitCase(s string, initCase bool) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	// Verifica se a string está na lista de acrônimos
	if a, hasAcronym := uppercaseAcronym.Load(s); hasAcronym {
		s = a.(string)
	}

	// Usa strings.Builder para performance
	n := strings.Builder{}
	n.Grow(len(s))

	capNext := initCase // Se deve capitalizar o próximo caractere
	prevIsCap := false  // Se o caractere anterior era maiúsculo

	for i, r := range s {
		vIsCap := unicode.IsUpper(r)
		vIsLow := unicode.IsLower(r)

		switch {
		case capNext:
			// Se deve capitalizar e o caractere é minúsculo, converte para maiúsculo
			if vIsLow {
				r = unicode.ToUpper(r)
			}
		case i == 0:
			// Se é o primeiro caractere e não deve começar com maiúscula
			if vIsCap {
				r = unicode.ToLower(r)
			}
		case prevIsCap && vIsCap && !strings.EqualFold(s, strings.ToUpper(s)):
			// Se o caractere anterior e o atual são maiúsculos e a string não é um acrônimo completo
			// converte para minúsculo para evitar sequências como "ABCdef"
			r = unicode.ToLower(r)
		}

		prevIsCap = vIsCap

		if vIsCap || vIsLow {
			// Se é uma letra, adiciona à string de saída
			n.WriteRune(r)
			capNext = false
		} else if unicode.IsDigit(r) {
			// Se é um dígito, adiciona e configura para capitalizar o próximo
			n.WriteRune(r)
			capNext = true
		} else if r == '_' || r == ' ' || r == '-' || r == '.' {
			// Se é um separador, configura para capitalizar o próximo
			capNext = true
		}
		// Ignora outros caracteres
	}
	return n.String()
}

// ToCamel converte uma string para CamelCase.
// Exemplos:
//
//	"test_case" -> "TestCase"
//	"test.case" -> "TestCase"
//	"test" -> "Test"
//	"TestCase" -> "TestCase"
//	" test  case " -> "TestCase"
//	"" -> ""
//	"many_many_words" -> "ManyManyWords"
//	"AnyKind of_string" -> "AnyKindOfString"
//	"odd-fix" -> "OddFix"
//	"numbers2And55with000" -> "Numbers2And55With000"
func ToCamel(s string) string {
	return toCamelInitCase(s, true)
}

// ToLowerCamel converte uma string para lowerCamelCase.
// Exemplos:
//
//	"foo-bar" -> "fooBar"
//	"TestCase" -> "testCase"
//	"" -> ""
//	"AnyKind of_string" -> "anyKindOfString"
//	"AnyKind.of-string" -> "anyKindOfString"
//	"ID" -> "id" (a menos que configurado com ConfigureAcronym)
//	"some string" -> "someString"
//	" some string" -> "someString"
func ToLowerCamel(s string) string {
	return toCamelInitCase(s, false)
}

// ToSnake converte uma string para snake_case.
// Exemplos:
//
//	"TestCase" -> "test_case"
//	"Test Case" -> "test_case"
//	"TestCase" -> "test_case"
//	"Test-Case" -> "test_case"
//	"test-case" -> "test_case"
//	"test_case" -> "test_case"
//	"TEST_CASE" -> "test_case"
func ToSnake(s string) string {
	return toSeparatedCase(s, '_', false)
}

// ToScreamingSnake converte uma string para SCREAMING_SNAKE_CASE (todas maiúsculas com underscores).
// Exemplos:
//
//	"TestCase" -> "TEST_CASE"
//	"Test Case" -> "TEST_CASE"
//	"TestCase" -> "TEST_CASE"
//	"Test-Case" -> "TEST_CASE"
//	"test-case" -> "TEST_CASE"
//	"test_case" -> "TEST_CASE"
//	"TEST_CASE" -> "TEST_CASE"
func ToScreamingSnake(s string) string {
	return toSeparatedCase(s, '_', true)
}

// ToKebab converte uma string para kebab-case.
// Exemplos:
//
//	"TestCase" -> "test-case"
//	"Test Case" -> "test-case"
//	"TestCase" -> "test-case"
//	"Test-Case" -> "test-case"
//	"test-case" -> "test-case"
//	"test_case" -> "test-case"
//	"TEST_CASE" -> "test-case"
func ToKebab(s string) string {
	return toSeparatedCase(s, '-', false)
}

// ToScreamingKebab converte uma string para SCREAMING-KEBAB-CASE (todas maiúsculas com hífens).
// Exemplos:
//
//	"TestCase" -> "TEST-CASE"
//	"Test Case" -> "TEST-CASE"
//	"TestCase" -> "TEST-CASE"
//	"Test-Case" -> "TEST-CASE"
//	"test-case" -> "TEST-CASE"
//	"test_case" -> "TEST-CASE"
//	"TEST_CASE" -> "TEST-CASE"
func ToScreamingKebab(s string) string {
	return toSeparatedCase(s, '-', true)
}

// toSeparatedCase é uma função auxiliar que converte uma string para um formato
// separado por um caractere específico (como snake_case ou kebab-case).
func toSeparatedCase(s string, separator rune, screaming bool) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	n := strings.Builder{}
	n.Grow(len(s) + 2) // crescimento conservador para acomodar separadores adicionais

	var prevIsLower, prevIsUpper, currIsUpper, prevIsNumber, currIsNumber bool
	var lastRune rune

	addSeparator := func() {
		if n.Len() > 0 && lastRune != separator {
			n.WriteRune(separator)
			lastRune = separator
		}
	}

	for i, r := range s {
		currIsUpper = unicode.IsUpper(r)
		currIsNumber = unicode.IsDigit(r)

		if r == '_' || r == ' ' || r == '-' || r == '.' {
			// É um separador
			addSeparator()
			continue
		}

		// Adiciona um separador quando:
		// 1. Estamos mudando de minúsculo para maiúsculo (exceptoCase -> excepto_case)
		// 2. Estamos mudando de número para letra (http2ssl -> http2_ssl)
		// 3. Estamos mudando de maiúsculo para minúsculo, mas apenas depois da segunda letra (SSLConn -> SSL_Conn -> ssl_conn)
		if i > 0 &&
			((prevIsLower && currIsUpper) ||
				(prevIsNumber && !currIsNumber) ||
				(prevIsUpper && currIsUpper && i+1 < len(s) && unicode.IsLower([]rune(s)[i+1]))) {
			addSeparator()
		}

		if screaming {
			r = unicode.ToUpper(r)
		} else {
			r = unicode.ToLower(r)
		}

		n.WriteRune(r)
		lastRune = r

		prevIsLower = unicode.IsLower(r)
		prevIsUpper = currIsUpper
		prevIsNumber = currIsNumber
	}

	return n.String()
}
