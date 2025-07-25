package checks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMigrationCompleteness verifica se todos os checks do _old/validator foram migrados
func TestMigrationCompleteness(t *testing.T) {
	t.Run("DateTime checks migrated", func(t *testing.T) {
		// DateTimeChecker migrado para DateTimeFormatChecker
		checker := NewDateTimeFormatChecker()
		assert.NotNil(t, checker)

		// Teste com os mesmos dados do antigo
		assert.True(t, checker.IsFormat("2006-01-02"))
		assert.True(t, checker.IsFormat("2006-01-02T15:04:05Z"))
		assert.False(t, checker.IsFormat("invalid-date"))
	})

	t.Run("EmptyStringChecker migrated", func(t *testing.T) {
		// EmptyStringChecker migrado para NonEmptyStringChecker
		checker := NewNonEmptyStringChecker()
		assert.NotNil(t, checker)

		// Teste de comportamento oposto ao antigo (antigo retornava false para empty)
		assert.False(t, checker.IsFormat(""))
		assert.True(t, checker.IsFormat("test"))
		assert.False(t, checker.IsFormat(123))
	})

	t.Run("Iso8601Date migrated", func(t *testing.T) {
		// Iso8601Date migrado para ISO8601DateChecker
		checker := NewISO8601DateChecker()
		assert.NotNil(t, checker)

		assert.True(t, checker.IsFormat("2006-01-02T15:04:05-07:00"))
		assert.False(t, checker.IsFormat("invalid-date"))
	})

	t.Run("JsonNumber migrated", func(t *testing.T) {
		// JsonNumber migrado para JSONNumberChecker
		checker := NewJSONNumberChecker()
		assert.NotNil(t, checker)

		// Mesma funcionalidade
		assert.True(t, checker.IsFormat(json.Number("123")))
		assert.False(t, checker.IsFormat(123))
	})

	t.Run("Decimal migrated", func(t *testing.T) {
		// Decimal migrado para DecimalChecker
		checker := NewDecimalChecker()
		factor8Checker := NewDecimalCheckerByFactor8()

		assert.NotNil(t, checker)
		assert.NotNil(t, factor8Checker)

		// Funcionalidade melhorada sem dependência externa
		assert.True(t, checker.IsFormat(123.45))
		assert.True(t, factor8Checker.IsFormat(123.45))
	})

	t.Run("StrongNameFormat migrated", func(t *testing.T) {
		// StrongNameFormat migrado para StrongNameFormatChecker
		checker := NewStrongNameFormatChecker()
		assert.NotNil(t, checker)

		// Mesma funcionalidade
		assert.True(t, checker.IsFormat("VALID_NAME"))
		assert.True(t, checker.IsFormat("_UNDER_SCORE"))
		assert.False(t, checker.IsFormat("lowercase"))
		assert.False(t, checker.IsFormat("123"))
	})

	t.Run("TextMatch checks migrated", func(t *testing.T) {
		// TextMatch migrado para TextMatchChecker
		textChecker := NewTextMatchChecker()
		textWithNumberChecker := NewTextMatchWithNumberChecker()
		customChecker, err := NewCustomRegexChecker("^[a-z]+@[a-z]+\\.[a-z]{2,}$")
		assert.NoError(t, err)

		assert.NotNil(t, textChecker)
		assert.NotNil(t, textWithNumberChecker)
		assert.NotNil(t, customChecker)

		// TextMatch: apenas letras, espaços e underscore
		assert.True(t, textChecker.IsFormat("valid text"))
		assert.False(t, textChecker.IsFormat("invalid123"))

		// TextMatchWithNumber: inclui números
		assert.True(t, textWithNumberChecker.IsFormat("valid text 123"))
		assert.False(t, textWithNumberChecker.IsFormat("invalid@"))

		// TextMatchCustom: regex customizada
		assert.True(t, customChecker.IsFormat("test@example.com"))
		assert.False(t, customChecker.IsFormat("invalid-email"))
	})
}
