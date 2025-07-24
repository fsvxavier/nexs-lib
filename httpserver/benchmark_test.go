package httpserver

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// BenchmarkRegistryRegister benchmarks server factory registration
func BenchmarkRegistryRegister(b *testing.B) {
	registry := NewRegistry()
	factory := func(cfg interface{}) (interfaces.HTTPServer, error) {
		return &mockServer{}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := "server" + string(rune(i%1000))
		registry.Register(name, factory)
	}
}

// BenchmarkRegistryCreate benchmarks server creation
func BenchmarkRegistryCreate(b *testing.B) {
	registry := NewRegistry()
	factory := func(cfg interface{}) (interfaces.HTTPServer, error) {
		return &mockServer{}, nil
	}
	registry.Register("test", factory)
	cfg := config.DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server, err := registry.Create("test", cfg)
		if err != nil {
			b.Fatal(err)
		}
		_ = server
	}
}

// BenchmarkObservableServerOperations benchmarks observable server operations
func BenchmarkObservableServerOperations(b *testing.B) {
	registry := NewRegistry()
	factory := func(cfg interface{}) (interfaces.HTTPServer, error) {
		return &mockServer{}, nil
	}
	registry.Register("test", factory)

	// Attach multiple observers
	for i := 0; i < 5; i++ {
		registry.AttachObserver(&benchmarkObserver{})
	}

	cfg := config.DefaultConfig()
	server, _ := registry.Create("test", cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.Start()
		server.Stop(context.Background())
	}
}

// BenchmarkRegistryWithObservers benchmarks registry with multiple observers
func BenchmarkRegistryWithObservers(b *testing.B) {
	registry := NewRegistry()
	factory := func(cfg interface{}) (interfaces.HTTPServer, error) {
		return &mockServer{}, nil
	}
	registry.Register("test", factory)

	// Attach many observers
	for i := 0; i < 100; i++ {
		registry.AttachObserver(&benchmarkObserver{})
	}

	cfg := config.DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server, err := registry.Create("test", cfg)
		if err != nil {
			b.Fatal(err)
		}
		server.Start()
		server.Stop(context.Background())
	}
}

// BenchmarkSetHandler benchmarks handler setting
func BenchmarkSetHandler(b *testing.B) {
	registry := NewRegistry()
	factory := func(cfg interface{}) (interfaces.HTTPServer, error) {
		return &mockServer{}, nil
	}
	registry.Register("test", factory)

	cfg := config.DefaultConfig()
	server, _ := registry.Create("test", cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.SetHandler(handler)
	}
}

// BenchmarkConcurrentRegistryOperations benchmarks concurrent operations
func BenchmarkConcurrentRegistryOperations(b *testing.B) {
	registry := NewRegistry()
	factory := func(cfg interface{}) (interfaces.HTTPServer, error) {
		return &mockServer{}, nil
	}

	// Pre-register some servers
	for i := 0; i < 10; i++ {
		name := "server" + string(rune('0'+i))
		registry.Register(name, factory)
	}

	cfg := config.DefaultConfig()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			server, err := registry.Create("server0", cfg)
			if err != nil {
				b.Fatal(err)
			}
			_ = server
		}
	})
} // benchmarkObserver for benchmarking
type benchmarkObserver struct{}

func (m *benchmarkObserver) OnStart(name string) {}
func (m *benchmarkObserver) OnStop(name string)  {}
func (m *benchmarkObserver) OnRequest(name string, req *http.Request, status int, duration time.Duration) {
}
func (m *benchmarkObserver) OnBeforeRequest(name string, req *http.Request) {}
func (m *benchmarkObserver) OnAfterRequest(name string, req *http.Request, status int, duration time.Duration) {
}
