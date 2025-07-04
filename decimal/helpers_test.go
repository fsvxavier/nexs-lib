package decimal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelpers(t *testing.T) {
	t.Run("NewDecimal", func(t *testing.T) {
		decimal, err := NewDecimal("123.456")
		assert.NoError(t, err)
		assert.Equal(t, "123.456", decimal.String())
	})
	t.Run("NewDecimalWithProvider", func(t *testing.T) {
		// ShopSpring
		decimal, err := NewDecimalWithProvider("123.456", ShopSpring)
		assert.NoError(t, err)
		assert.Equal(t, "123.456", decimal.String())

		// APD
		decimal, err = NewDecimalWithProvider("123.456", APD)
		assert.NoError(t, err)
		assert.Equal(t, "123.456", decimal.String())
	})

	t.Run("ShopSpringDecimal", func(t *testing.T) {
		decimal, err := ShopSpringDecimal("123.456")
		assert.NoError(t, err)
		assert.Equal(t, "123.456", decimal.String())
	})

	t.Run("APDDecimal", func(t *testing.T) {
		decimal, err := APDDecimal("123.456")
		assert.NoError(t, err)
		assert.Equal(t, "123.456", decimal.String())
	})
}
