// Package main demonstrates compression middleware usage.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/middleware"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/compression"
)

func main() {
	// Create middleware chain
	chain := middleware.NewChain()

	// Configure compression middleware
	compressionConfig := compression.Config{
		Enabled:   true,
		SkipPaths: []string{"/health", "/small"}, // Skip small responses
		Level:     6,                             // Balanced compression level
		MinSize:   1024,                          // Only compress responses larger than 1KB
		Types: []string{
			"text/html",
			"text/css",
			"text/javascript",
			"text/plain",
			"application/json",
			"application/javascript",
			"application/xml",
			"application/rss+xml",
			"application/atom+xml",
			"image/svg+xml",
		},
	}

	// Add compression middleware
	chain.Add(compression.NewMiddleware(compressionConfig))

	// Setup routes
	mux := http.NewServeMux()

	// Health endpoint (not compressed)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Small response (not compressed due to MinSize)
	mux.HandleFunc("/small", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Small response",
		})
	})

	// Large JSON response (will be compressed)
	largeJSONHandler := chain.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Generate a large JSON response
		data := make([]map[string]interface{}, 0, 1000)
		for i := 0; i < 1000; i++ {
			item := map[string]interface{}{
				"id":          i,
				"name":        fmt.Sprintf("Item %d", i),
				"description": fmt.Sprintf("This is a detailed description for item number %d. It contains a lot of text to make the response larger and demonstrate compression benefits.", i),
				"timestamp":   time.Now().Add(time.Duration(i) * time.Minute).Format(time.RFC3339),
				"metadata": map[string]interface{}{
					"category":    "sample",
					"subcategory": fmt.Sprintf("sub-%d", i%10),
					"tags":        []string{"tag1", "tag2", "tag3"},
					"active":      i%2 == 0,
				},
			}
			data = append(data, item)
		}

		response := map[string]interface{}{
			"message": "Large JSON response with compression",
			"count":   len(data),
			"data":    data,
		}

		json.NewEncoder(w).Encode(response)
	}))

	mux.Handle("/api/large", largeJSONHandler)

	// Large HTML response (will be compressed)
	largeHTMLHandler := chain.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Compression Demo</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .item { border: 1px solid #ddd; padding: 20px; margin: 10px 0; }
        .description { color: #666; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Compression Middleware Demo</h1>
        <p>This page demonstrates HTTP response compression. The content below is generated dynamically to create a large HTML response that benefits from compression.</p>
`

		// Add repetitive content to make the response larger
		for i := 0; i < 100; i++ {
			html += fmt.Sprintf(`
        <div class="item">
            <h3>Item %d</h3>
            <p class="description">This is item number %d. It contains repetitive content that compresses very well using gzip or deflate algorithms. The compression middleware automatically detects that this is an HTML response and applies compression based on the Accept-Encoding header sent by the client.</p>
        </div>`, i, i)
		}

		html += `
    </div>
    <script>
        console.log('Page loaded with compression middleware');
        // This JavaScript will also be compressed
        for (let i = 0; i < 1000; i++) {
            console.log('Repetitive log message number ' + i + ' to increase response size');
        }
    </script>
</body>
</html>`

		w.Write([]byte(html))
	}))

	mux.Handle("/page", largeHTMLHandler)

	// Text response (will be compressed)
	largeTextHandler := chain.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		// Generate large text content
		var content strings.Builder
		content.WriteString("Compression Middleware Demo - Large Text Response\n")
		content.WriteString("=" + strings.Repeat("=", 50) + "\n\n")

		for i := 0; i < 500; i++ {
			content.WriteString(fmt.Sprintf("Line %d: This is a repetitive line of text that demonstrates how well text content compresses. The compression middleware will automatically detect the Content-Type and apply the appropriate compression algorithm.\n", i))
		}

		w.Write([]byte(content.String()))
	}))

	mux.Handle("/text", largeTextHandler)

	// Show compression info
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		acceptEncoding := r.Header.Get("Accept-Encoding")
		userAgent := r.Header.Get("User-Agent")

		response := map[string]interface{}{
			"compression_info": map[string]interface{}{
				"client_accepts": acceptEncoding,
				"user_agent":     userAgent,
				"supported":      []string{"gzip", "deflate"},
				"min_size":       1024,
				"level":          6,
			},
			"endpoints": map[string]interface{}{
				"/health":    "Not compressed (skipped path)",
				"/small":     "Not compressed (too small)",
				"/api/large": "Compressed JSON (large response)",
				"/page":      "Compressed HTML (large response)",
				"/text":      "Compressed text (large response)",
			},
		}

		json.NewEncoder(w).Encode(response)
	})

	fmt.Println("Compression server starting on :8080")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET /health       - Health check (not compressed)")
	fmt.Println("  GET /small        - Small response (not compressed)")
	fmt.Println("  GET /api/large    - Large JSON (compressed)")
	fmt.Println("  GET /page         - Large HTML (compressed)")
	fmt.Println("  GET /text         - Large text (compressed)")
	fmt.Println("  GET /info         - Compression info")
	fmt.Println("")
	fmt.Println("Test with different Accept-Encoding headers:")
	fmt.Println("  curl -H 'Accept-Encoding: gzip' http://localhost:8080/api/large")
	fmt.Println("  curl -H 'Accept-Encoding: deflate' http://localhost:8080/page")
	fmt.Println("  curl -H 'Accept-Encoding: gzip,deflate' http://localhost:8080/text")
	fmt.Println("  curl http://localhost:8080/api/large  # No compression")
	fmt.Println("")
	fmt.Println("Check response headers for Content-Encoding and Vary headers")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
