package paginate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockHttpRequest implementa a interface HttpRequest para testes
type MockHttpRequest struct {
	params map[string]string
}

func (m *MockHttpRequest) Query(key string) string {
	if val, ok := m.params[key]; ok {
		return val
	}
	return ""
}

func (m *MockHttpRequest) QueryParam(key string) string {
	return m.Query(key)
}

func NewMockRequest(params map[string]string) *MockHttpRequest {
	return &MockHttpRequest{
		params: params,
	}
}

func TestNewMetadata(t *testing.T) {
	// Teste com valores padrão
	metadata := NewMetadata()
	assert.Equal(t, 1, metadata.Page.CurrentPage)
	assert.Equal(t, 150, metadata.Page.RecordsPerPage)
	assert.Equal(t, "id", metadata.Sort.Field)
	assert.Equal(t, "asc", metadata.Sort.Order)

	// Teste com opções específicas
	metadata = NewMetadata(
		WithPage(5),
		WithLimit(20),
		WithSort("name"),
		WithOrder("desc"),
	)
	assert.Equal(t, 5, metadata.Page.CurrentPage)
	assert.Equal(t, 20, metadata.Page.RecordsPerPage)
	assert.Equal(t, "name", metadata.Sort.Field)
	assert.Equal(t, "desc", metadata.Sort.Order)
}

func TestPage_CalculateNextPreviousPage(t *testing.T) {
	tests := []struct {
		name           string
		currentPage    int
		totalData      int
		recordsPerPage int
		wantPrevious   int
		wantNext       int
		wantTotalPages int
	}{
		{
			name:           "Primeira página com mais páginas",
			currentPage:    1,
			totalData:      100,
			recordsPerPage: 10,
			wantPrevious:   0,
			wantNext:       2,
			wantTotalPages: 10,
		},
		{
			name:           "Página intermediária",
			currentPage:    5,
			totalData:      100,
			recordsPerPage: 10,
			wantPrevious:   4,
			wantNext:       6,
			wantTotalPages: 10,
		},
		{
			name:           "Última página",
			currentPage:    10,
			totalData:      100,
			recordsPerPage: 10,
			wantPrevious:   9,
			wantNext:       0,
			wantTotalPages: 10,
		},
		{
			name:           "Sem dados",
			currentPage:    1,
			totalData:      0,
			recordsPerPage: 10,
			wantPrevious:   0,
			wantNext:       0,
			wantTotalPages: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Page{
				CurrentPage:    tt.currentPage,
				RecordsPerPage: tt.recordsPerPage,
			}
			totalData := tt.totalData
			p.CalculateNextPreviousPage(totalData)

			assert.Equal(t, tt.wantPrevious, p.Previous)
			assert.Equal(t, tt.wantNext, p.Next)
			assert.Equal(t, tt.wantTotalPages, p.TotalPages)
		})
	}
}

func TestNewOutput(t *testing.T) {
	// Conteúdo para teste
	content := []map[string]interface{}{
		{"id": 1, "name": "Teste 1"},
		{"id": 2, "name": "Teste 2"},
	}

	// Metadados para teste
	metadata := NewMetadata(
		WithPage(1),
		WithLimit(10),
		WithSort("id"),
		WithOrder("asc"),
	)

	// Criar saída
	output := NewOutput(content, metadata)

	// Verificar se o conteúdo foi corretamente atribuído
	contentSlice, ok := output.Content.([]map[string]interface{})
	assert.True(t, ok, "Conteúdo deveria ser uma slice de maps")
	assert.Len(t, contentSlice, 2, "Deveria ter 2 itens no conteúdo")

	// Verificar metadados
	assert.Equal(t, 1, output.Metadata.Page.CurrentPage)
	assert.Equal(t, 10, output.Metadata.Page.RecordsPerPage)
	assert.Equal(t, "id", output.Metadata.Sort.Field)
	assert.Equal(t, "asc", output.Metadata.Sort.Order)
}

func TestNewOutputWithTotal(t *testing.T) {
	ctx := context.Background()

	// Conteúdo para teste
	content := []map[string]interface{}{
		{"id": 1, "name": "Teste 1"},
		{"id": 2, "name": "Teste 2"},
	}

	// Metadados para teste
	metadata := NewMetadata(
		WithPage(1),
		WithLimit(10),
		WithSort("id"),
		WithOrder("asc"),
	)

	// Total de dados
	totalData := 20

	// Criar saída com total
	output, err := NewOutputWithTotal(ctx, content, totalData, metadata)
	assert.NoError(t, err)

	// Verificar se o conteúdo foi corretamente atribuído
	contentSlice, ok := output.Content.([]map[string]interface{})
	assert.True(t, ok, "Conteúdo deveria ser uma slice de maps")
	assert.Len(t, contentSlice, 2, "Deveria ter 2 itens no conteúdo")

	// Verificar metadados
	assert.Equal(t, 1, output.Metadata.Page.CurrentPage)
	assert.Equal(t, 10, output.Metadata.Page.RecordsPerPage)
	assert.Equal(t, "id", output.Metadata.Sort.Field)
	assert.Equal(t, "asc", output.Metadata.Sort.Order)
	assert.Equal(t, 0, output.Metadata.Page.Previous)
	assert.Equal(t, 2, output.Metadata.Page.Next)
	assert.Equal(t, 2, output.Metadata.Page.TotalPages)
	assert.Equal(t, 20, output.Metadata.TotalData)

	// Testar erro quando a página atual é maior que o total de páginas
	metadata = NewMetadata(WithPage(10), WithLimit(10))
	output, err = NewOutputWithTotal(ctx, content, 20, metadata)
	assert.Error(t, err, "Deveria retornar erro quando a página atual é maior que o total de páginas")
}

func TestPreparePagination(t *testing.T) {
	metadata := NewMetadata(
		WithPage(2),
		WithLimit(15),
		WithSort("name"),
		WithOrder("desc"),
		WithQuery("SELECT * FROM users WHERE active = true"),
	)

	// Preparar paginação
	result := metadata.PreparePagination()

	// Verificar query resultante
	expectedQuery := "SELECT * FROM users WHERE active = true ORDER BY name desc LIMIT 15 OFFSET 15"
	assert.Equal(t, expectedQuery, result.Query)

	// Verificar se o mesmo objeto foi retornado
	assert.Same(t, metadata, result)
}

func TestPaginationIndices(t *testing.T) {
	tests := []struct {
		name           string
		currentPage    int
		recordsPerPage int
		totalItems     int
		wantStartIndex int
		wantEndIndex   int
		wantValid      bool
	}{
		{
			name:           "Primeira página com dados",
			currentPage:    1,
			recordsPerPage: 10,
			totalItems:     100,
			wantStartIndex: 0,
			wantEndIndex:   10,
			wantValid:      true,
		},
		{
			name:           "Segunda página com dados",
			currentPage:    2,
			recordsPerPage: 10,
			totalItems:     100,
			wantStartIndex: 10,
			wantEndIndex:   20,
			wantValid:      true,
		},
		{
			name:           "Última página com menos itens",
			currentPage:    10,
			recordsPerPage: 10,
			totalItems:     95,
			wantStartIndex: 90,
			wantEndIndex:   95,
			wantValid:      true,
		},
		{
			name:           "Página além do total",
			currentPage:    12,
			recordsPerPage: 10,
			totalItems:     95,
			wantStartIndex: 0,
			wantEndIndex:   0,
			wantValid:      false,
		},
		{
			name:           "Lista vazia",
			currentPage:    1,
			recordsPerPage: 10,
			totalItems:     0,
			wantStartIndex: 0,
			wantEndIndex:   0,
			wantValid:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := NewMetadata(
				WithPage(tt.currentPage),
				WithLimit(tt.recordsPerPage),
			)

			startIndex, endIndex, isValid := PaginationIndices(metadata, tt.totalItems)

			assert.Equal(t, tt.wantStartIndex, startIndex)
			assert.Equal(t, tt.wantEndIndex, endIndex)
			assert.Equal(t, tt.wantValid, isValid)
		})
	}
}

func TestApplyPaginationToSlice(t *testing.T) {
	ctx := context.Background()

	// Criar dados de teste
	testData := []map[string]interface{}{
		{"id": 1, "name": "Item 1"},
		{"id": 2, "name": "Item 2"},
		{"id": 3, "name": "Item 3"},
		{"id": 4, "name": "Item 4"},
		{"id": 5, "name": "Item 5"},
		{"id": 6, "name": "Item 6"},
		{"id": 7, "name": "Item 7"},
		{"id": 8, "name": "Item 8"},
		{"id": 9, "name": "Item 9"},
		{"id": 10, "name": "Item 10"},
		{"id": 11, "name": "Item 11"},
		{"id": 12, "name": "Item 12"},
	}

	tests := []struct {
		name           string
		currentPage    int
		recordsPerPage int
		wantItems      int
		wantError      bool
	}{
		{
			name:           "Primeira página",
			currentPage:    1,
			recordsPerPage: 5,
			wantItems:      5,
			wantError:      false,
		},
		{
			name:           "Segunda página",
			currentPage:    2,
			recordsPerPage: 5,
			wantItems:      5,
			wantError:      false,
		},
		{
			name:           "Última página com menos itens",
			currentPage:    3,
			recordsPerPage: 5,
			wantItems:      2,
			wantError:      false,
		},
		{
			name:           "Página além do total",
			currentPage:    4,
			recordsPerPage: 5,
			wantItems:      0,
			wantError:      true,
		},
		{
			name:           "Lista vazia",
			currentPage:    1,
			recordsPerPage: 5,
			wantItems:      0,
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := NewMetadata(
				WithPage(tt.currentPage),
				WithLimit(tt.recordsPerPage),
			)

			var data []map[string]interface{}
			if tt.name != "Lista vazia" {
				data = testData
			} else {
				data = []map[string]interface{}{} // Lista vazia mas inicializada
			}

			output, err := ApplyPaginationToSlice(ctx, data, metadata)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			content, ok := output.Content.([]interface{})
			assert.True(t, ok)
			assert.Len(t, content, tt.wantItems)

			// Verificar se os metadados foram atualizados corretamente
			if !tt.wantError && tt.name != "Lista vazia" {
				assert.Equal(t, len(testData), output.Metadata.TotalData)
				totalPages := (len(testData) + tt.recordsPerPage - 1) / tt.recordsPerPage
				assert.Equal(t, totalPages, output.Metadata.Page.TotalPages)
			}
		})
	}
}

func TestToInterfaceSlice(t *testing.T) {
	t.Run("[]interface{}", func(t *testing.T) {
		input := []interface{}{1, "dois", 3.0}
		result, ok := toInterfaceSlice(input)
		assert.True(t, ok)
		assert.Equal(t, input, result)
	})

	t.Run("[]map[string]interface{}", func(t *testing.T) {
		input := []map[string]interface{}{
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"},
		}
		result, ok := toInterfaceSlice(input)
		assert.True(t, ok)
		assert.Len(t, result, 2)
	})

	t.Run("[]string", func(t *testing.T) {
		input := []string{"um", "dois", "três"}
		result, ok := toInterfaceSlice(input)
		assert.True(t, ok)
		assert.Len(t, result, 3)
		assert.Equal(t, "um", result[0])
	})

	t.Run("[]int", func(t *testing.T) {
		input := []int{1, 2, 3}
		result, ok := toInterfaceSlice(input)
		assert.True(t, ok)
		assert.Len(t, result, 3)
		assert.Equal(t, 1, result[0])
	})

	t.Run("struct slice", func(t *testing.T) {
		type TestStruct struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		input := []TestStruct{
			{ID: 1, Name: "Test 1"},
			{ID: 2, Name: "Test 2"},
		}
		result, ok := toInterfaceSlice(input)
		assert.True(t, ok)
		assert.Len(t, result, 2)
	})

	t.Run("non-slice", func(t *testing.T) {
		input := "not a slice"
		result, ok := toInterfaceSlice(input)
		assert.False(t, ok)
		assert.Nil(t, result)
	})
}
