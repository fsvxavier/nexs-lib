package paginate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFromRequest(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		params    map[string]string
		sortable  []string
		wantErr   bool
		wantPage  int
		wantLimit int
		wantSort  string
		wantOrder string
	}{
		{
			name: "Parâmetros válidos completos",
			params: map[string]string{
				"page":  "2",
				"limit": "20",
				"sort":  "name",
				"order": "desc",
			},
			sortable:  []string{"id", "name", "created_at"},
			wantErr:   false,
			wantPage:  2,
			wantLimit: 20,
			wantSort:  "name",
			wantOrder: "desc",
		},
		{
			name: "Parâmetros parciais",
			params: map[string]string{
				"page": "3",
			},
			sortable:  []string{"id", "name"},
			wantErr:   false,
			wantPage:  3,
			wantLimit: 150,   // Valor padrão
			wantSort:  "id",  // Valor padrão
			wantOrder: "asc", // Valor padrão
		},
		{
			name: "Página inválida",
			params: map[string]string{
				"page": "-1",
			},
			sortable: []string{"id", "name"},
			wantErr:  true,
		},
		{
			name:      "Sem parâmetros",
			params:    map[string]string{},
			sortable:  []string{"id", "name"},
			wantErr:   false,
			wantPage:  1,     // Valor padrão
			wantLimit: 150,   // Valor padrão
			wantSort:  "id",  // Valor padrão
			wantOrder: "asc", // Valor padrão
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewMockRequest(tt.params)
			metadata, err := ParseFromRequest(ctx, req, tt.sortable...)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPage, metadata.Page.CurrentPage)
				assert.Equal(t, tt.wantLimit, metadata.Page.RecordsPerPage)
				assert.Equal(t, tt.wantSort, metadata.Sort.Field)
				assert.Equal(t, tt.wantOrder, metadata.Sort.Order)
			}
		})
	}
}

// MockRequest implementa a interface HttpRequest para testes
type MockRequest struct {
	params map[string]string
}

func (r *MockRequest) Query(key string) string {
	return r.params[key]
}

func (r *MockRequest) QueryParam(key string) string {
	return r.params[key]
}

func TestSet_Add(t *testing.T) {
	set := Set[string]{}
	set.Add("test")

	assert.True(t, set.Contains("test"))
}

func TestSet_Remove(t *testing.T) {
	set := Set[string]{}
	set.Add("test")
	set.Remove("test")

	assert.False(t, set.Contains("test"))
}

func TestSet(t *testing.T) {
	t.Run("Add e Contains", func(t *testing.T) {
		set := Set[string]{}

		// Verifica que elemento não existe inicialmente
		assert.False(t, set.Contains("test"))

		// Adiciona e verifica
		set.Add("test")
		assert.True(t, set.Contains("test"))

		// Adiciona outro elemento
		set.Add("another")
		assert.True(t, set.Contains("test"))
		assert.True(t, set.Contains("another"))
	})

	t.Run("Remove", func(t *testing.T) {
		set := Set[int]{}

		// Adiciona e verifica
		set.Add(1)
		set.Add(2)
		assert.True(t, set.Contains(1))
		assert.True(t, set.Contains(2))

		// Remove e verifica
		set.Remove(1)
		assert.False(t, set.Contains(1))
		assert.True(t, set.Contains(2))

		// Remove elemento que não existe não deve causar erro
		set.Remove(3)
	})

	t.Run("Diferentes tipos", func(t *testing.T) {
		// Teste com diferentes tipos para garantir que o genérico funciona
		intSet := Set[int]{}
		intSet.Add(42)
		assert.True(t, intSet.Contains(42))

		floatSet := Set[float64]{}
		floatSet.Add(3.14)
		assert.True(t, floatSet.Contains(3.14))

		// Usando tipo struct
		type Person struct {
			Name string
			Age  int
		}

		personSet := Set[Person]{}
		p1 := Person{Name: "Alice", Age: 30}
		p2 := Person{Name: "Bob", Age: 25}

		personSet.Add(p1)
		assert.True(t, personSet.Contains(p1))
		assert.False(t, personSet.Contains(p2))
	})
}
