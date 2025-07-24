package set_test

import (
	"testing"

	"github.com/dock-tech/isis-golang-lib/set"
	"github.com/stretchr/testify/assert"
)

func TestSet_Add(t *testing.T) {
	set := set.Set[string]{}
	set.Add("test")

	assert.True(t, set.Contains("test"))
}

func TestSet_Remove(t *testing.T) {
	set := set.Set[string]{}
	set.Add("test")
	set.Remove("test")

	assert.False(t, set.Contains("test"))
}

func TestSet_Contains(t *testing.T) {
	set := set.Set[string]{}
	set.Add("test")

	assert.True(t, set.Contains("test"))
	assert.False(t, set.Contains("not_in_set"))
}
