package pagination_test

import (
	"net/url"
	"testing"

	"github.com/fsvxavier/nexs-lib/pagination"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
)

func BenchmarkPaginationService_ParseRequest(b *testing.B) {
	service := pagination.NewPaginationService(nil)
	params := url.Values{
		"page":  []string{"2"},
		"limit": []string{"25"},
		"sort":  []string{"name"},
		"order": []string{"desc"},
	}
	sortableFields := []string{"id", "name", "created_at", "updated_at"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ParseRequest(params, sortableFields...)
	}
}

func BenchmarkPaginationService_BuildQuery(b *testing.B) {
	service := pagination.NewPaginationService(nil)
	baseQuery := "SELECT u.id, u.name, u.email, p.title FROM users u LEFT JOIN posts p ON u.id = p.user_id WHERE u.active = true"
	params := &interfaces.PaginationParams{
		Page:      3,
		Limit:     20,
		SortField: "u.created_at",
		SortOrder: "desc",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.BuildQuery(baseQuery, params)
	}
}

func BenchmarkPaginationService_CreateResponse(b *testing.B) {
	service := pagination.NewPaginationService(nil)
	content := make([]map[string]interface{}, 25)
	for i := 0; i < 25; i++ {
		content[i] = map[string]interface{}{
			"id":    i + 1,
			"name":  "User " + string(rune(i+65)),
			"email": "user" + string(rune(i+65)) + "@example.com",
		}
	}
	params := &interfaces.PaginationParams{
		Page:      2,
		Limit:     25,
		SortField: "name",
		SortOrder: "asc",
	}
	totalRecords := 125

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.CreateResponse(content, params, totalRecords)
	}
}

func BenchmarkPaginationService_BuildCountQuery(b *testing.B) {
	service := pagination.NewPaginationService(nil)
	baseQuery := "SELECT u.id, u.name, u.email, p.title FROM users u LEFT JOIN posts p ON u.id = p.user_id WHERE u.active = true"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.BuildCountQuery(baseQuery)
	}
}

func BenchmarkPaginationService_ValidatePageNumber(b *testing.B) {
	service := pagination.NewPaginationService(nil)
	params := &interfaces.PaginationParams{
		Page:  2,
		Limit: 25,
	}
	totalRecords := 125

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.ValidatePageNumber(params, totalRecords)
	}
}

// Benchmark different scenarios
func BenchmarkPaginationService_SmallDataset(b *testing.B) {
	service := pagination.NewPaginationService(nil)
	params := url.Values{
		"page":  []string{"1"},
		"limit": []string{"10"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsedParams, err := service.ParseRequest(params)
		if err != nil {
			b.Fatal(err)
		}
		_ = service.BuildQuery("SELECT * FROM users", parsedParams)
		_ = service.CreateResponse([]string{"item1", "item2"}, parsedParams, 50)
	}
}

func BenchmarkPaginationService_LargeDataset(b *testing.B) {
	service := pagination.NewPaginationService(nil)
	// Create large dataset mock
	largeContent := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		largeContent[i] = map[string]interface{}{
			"id":         i + 1,
			"field_A":    "value_" + string(rune(i%26+65)),
			"field_B":    i * 2,
			"field_C":    i * 3,
			"created_at": "2024-01-01",
		}
	}

	params := url.Values{
		"page":  []string{"100"},
		"limit": []string{"100"},
		"sort":  []string{"field_A"},
		"order": []string{"desc"},
	}

	sortableFields := []string{"id", "field_A", "field_B", "field_C", "created_at"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsedParams, err := service.ParseRequest(params, sortableFields...)
		if err != nil {
			b.Fatal(err)
		}
		_ = service.BuildQuery("SELECT * FROM large_table", parsedParams)
		_ = service.CreateResponse(largeContent, parsedParams, 100000)
	}
}

func BenchmarkPaginationService_ComplexQuery(b *testing.B) {
	service := pagination.NewPaginationService(nil)
	complexQuery := `
		SELECT 
			u.id, u.name, u.email, u.created_at,
			p.title, p.content, p.published_at,
			c.name as category_name,
			COUNT(likes.id) as likes_count,
			COUNT(comments.id) as comments_count
		FROM users u
		LEFT JOIN posts p ON u.id = p.user_id
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN likes ON p.id = likes.post_id
		LEFT JOIN comments ON p.id = comments.post_id
		WHERE u.active = true 
		AND p.published = true
		AND p.created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
		GROUP BY u.id, p.id, c.id
	`
	params := &interfaces.PaginationParams{
		Page:      5,
		Limit:     50,
		SortField: "p.published_at",
		SortOrder: "desc",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.BuildQuery(complexQuery, params)
		_ = service.BuildCountQuery(complexQuery)
	}
}

// Memory allocation benchmarks
func BenchmarkPaginationService_MemoryAllocation(b *testing.B) {
	service := pagination.NewPaginationService(nil)
	params := url.Values{
		"page":  []string{"2"},
		"limit": []string{"25"},
		"sort":  []string{"name"},
		"order": []string{"desc"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsedParams, _ := service.ParseRequest(params, "id", "name", "email")
		response := service.CreateResponse([]string{"item1", "item2"}, parsedParams, 100)
		_ = response
	}
}
