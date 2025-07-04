package paginate

// Este arquivo fornece exemplos de uso dos diferentes adaptadores
// para implementação em uma aplicação real

/*
# Uso com Fiber

```go
import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/fsvxavier/nexs-lib/paginate"
	pagefiber "github.com/fsvxavier/nexs-lib/paginate/fiber"
)

func ConfigureRoutes(app *fiber.App, service *YourService) {
	app.Get("/api/resources", func(c *fiber.Ctx) error {
		ctx := context.Background()

		// Extrair parâmetros de paginação
		metadata, err := pagefiber.Parse(ctx, c, "id", "name", "created_at")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		// Obter recursos paginados
		result, err := service.ListResources(ctx, metadata)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Retornar resultado paginado
		return c.JSON(result)
	})
}
```

# Uso com Atreugo

```go
import (
	"context"

	"github.com/savsgio/atreugo/v11"
	"github.com/fsvxavier/nexs-lib/paginate"
	pageAtreugo "github.com/fsvxavier/nexs-lib/paginate/atreugo"
)

func ConfigureRoutes(server *atreugo.Atreugo, service *YourService) {
	server.GET("/api/resources", func(ctx *atreugo.RequestCtx) error {
		appCtx := context.Background()

		// Extrair parâmetros de paginação
		metadata, err := pageAtreugo.Parse(appCtx, ctx, "id", "name", "created_at")
		if err != nil {
			return ctx.JSONResponse(map[string]interface{}{"error": err.Error()}, 400)
		}

		// Obter recursos paginados
		result, err := service.ListResources(appCtx, metadata)
		if err != nil {
			return ctx.JSONResponse(map[string]interface{}{"error": err.Error()}, 500)
		}

		// Retornar resultado paginado
		return ctx.JSONResponse(result, 200)
	})
}
```

# Uso com FastHTTP

```go
import (
	"context"
	"encoding/json"

	"github.com/valyala/fasthttp"
	"github.com/fsvxavier/nexs-lib/paginate"
	pageFastHTTP "github.com/fsvxavier/nexs-lib/paginate/fasthttp"
)

func ResourcesHandler(service *YourService) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		appCtx := context.Background()

		// Extrair parâmetros de paginação
		metadata, err := pageFastHTTP.Parse(appCtx, ctx, "id", "name", "created_at")
		if err != nil {
			ctx.SetStatusCode(400)
			response := map[string]interface{}{"error": err.Error()}
			jsonResponse, _ := json.Marshal(response)
			ctx.SetBody(jsonResponse)
			return
		}

		// Obter recursos paginados
		result, err := service.ListResources(appCtx, metadata)
		if err != nil {
			ctx.SetStatusCode(500)
			response := map[string]interface{}{"error": err.Error()}
			jsonResponse, _ := json.Marshal(response)
			ctx.SetBody(jsonResponse)
			return
		}

		// Retornar resultado paginado
		jsonResponse, _ := json.Marshal(result)
		ctx.SetContentType("application/json")
		ctx.SetBody(jsonResponse)
	}
}

func ConfigureRoutes() {
	service := NewYourService()

	// Configurar rotas
	handler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())

		switch path {
		case "/api/resources":
			ResourcesHandler(service)(ctx)
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
	}

	// Iniciar servidor
	fasthttp.ListenAndServe(":8080", handler)
}
```

# Uso com Echo

```go
import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/fsvxavier/nexs-lib/paginate"
	pageecho "github.com/fsvxavier/nexs-lib/paginate/echo"
)

func ConfigureRoutes(e *echo.Echo, service *YourService) {
	e.GET("/api/resources", func(c echo.Context) error {
		ctx := context.Background()

		// Extrair parâmetros de paginação
		metadata, err := pageecho.Parse(ctx, c, "id", "name", "created_at")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		}

		// Obter recursos paginados
		result, err := service.ListResources(ctx, metadata)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		}

		// Retornar resultado paginado
		return c.JSON(http.StatusOK, result)
	})
}
```

# Uso com Gin

```go
import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/fsvxavier/nexs-lib/paginate"
	pagegin "github.com/fsvxavier/nexs-lib/paginate/gin"
)

func ConfigureRoutes(r *gin.Engine, service *YourService) {
	r.GET("/api/resources", func(c *gin.Context) {
		ctx := context.Background()

		// Extrair parâmetros de paginação
		metadata, err := pagegin.Parse(ctx, c, "id", "name", "created_at")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Obter recursos paginados
		result, err := service.ListResources(ctx, metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Retornar resultado paginado
		c.JSON(http.StatusOK, result)
	})
}
```

# Uso com net/http

```go
import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/fsvxavier/nexs-lib/paginate"
	pagenethttp "github.com/fsvxavier/nexs-lib/paginate/nethttp"
)

func ResourcesHandler(service *YourService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// Extrair parâmetros de paginação
		metadata, err := pagenethttp.Parse(ctx, r, "id", "name", "created_at")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}

		// Obter recursos paginados
		result, err := service.ListResources(ctx, metadata)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}

		// Retornar resultado paginado
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func ConfigureRoutes() {
	service := NewYourService()

	http.HandleFunc("/api/resources", ResourcesHandler(service))
	http.ListenAndServe(":8080", nil)
}
```

# Uso com a biblioteca httpservers (Agnóstico de framework)

```go
import (
	"context"
	"log"

	"github.com/fsvxavier/nexs-lib/httpservers"
	"github.com/fsvxavier/nexs-lib/httpservers/common"
	"github.com/fsvxavier/nexs-lib/paginate"
)

// ResourceService representa uma interface genérica de serviço
type ResourceService interface {
	ListResources(ctx context.Context, metadata *page.Metadata) (*page.Output, error)
}

// ResourceHandler é um manipulador agnóstico que pode ser usado com qualquer servidor
type ResourceHandler struct {
	service ResourceService
}

// NewResourceHandler cria um novo manipulador de recursos
func NewResourceHandler(service ResourceService) *ResourceHandler {
	return &ResourceHandler{
		service: service,
	}
}

// ConfigureServer configura um servidor HTTP agnóstico usando a biblioteca httpservers
func ConfigureServer(service ResourceService) {
	// Criar manipulador
	handler := NewResourceHandler(service)

	// Escolher o tipo de servidor (Fiber, Echo, FastHTTP, etc.)
	serverType := httpservers.ServerTypeFiber // Pode ser qualquer tipo suportado

	// Criar servidor com opções comuns
	server, err := httpservers.NewServer(
		serverType,
		common.WithPort("8080"),
		common.WithHost("0.0.0.0"),
		common.WithSwagger(true),
		common.WithMetrics(true),
	)
	if err != nil {
		log.Fatalf("Erro ao criar servidor: %v", err)
	}

	// Registrar rotas independentemente do framework
	// O tipo específico de servidor pode ser acessado para configuração adicional
	switch serverType {
	case httpservers.ServerTypeFiber:
		// Configurar rotas para Fiber
		handler.ConfigureFiberRoutes(server)
	case httpservers.ServerTypeEcho:
		// Configurar rotas para Echo
		handler.ConfigureEchoRoutes(server)
	case httpservers.ServerTypeGin:
		// Configurar rotas para Gin
		handler.ConfigureGinRoutes(server)
	case httpservers.ServerTypeFastHTTP:
		// Configurar rotas para FastHTTP
		handler.ConfigureFastHTTPRoutes(server)
	case httpservers.ServerTypeAtreugo:
		// Configurar rotas para Atreugo
		handler.ConfigureAtreugoRoutes(server)
	}

	// Iniciar o servidor (independente do tipo)
	if err := server.Start(); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

// Os métodos a seguir demonstram como configurar rotas para diferentes tipos de servidores
// enquanto mantém o código de negócios comum

// ConfigureFiberRoutes configura rotas para o Fiber
func (h *ResourceHandler) ConfigureFiberRoutes(server interface{}) {
	// Importações específicas do Fiber
	import (
		"github.com/gofiber/fiber/v2"
		"github.com/fsvxavier/nexs-lib/httpservers/fiber"
		pagefiber "github.com/fsvxavier/nexs-lib/paginate/fiber"
	)

	// Converter para o tipo específico
	fiberServer := server.(*fiber.FiberServer)
	app := fiberServer.App()

	// Configurar rotas
	app.Get("/api/resources", func(c *fiber.Ctx) error {
		ctx := context.Background()

		// Extrair parâmetros de paginação
		metadata, err := pagefiber.Parse(ctx, c, "id", "name", "created_at")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		// Usar o serviço comum
		result, err := h.service.ListResources(ctx, metadata)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(result)
	})
}

// ConfigureEchoRoutes configura rotas para o Echo
func (h *ResourceHandler) ConfigureEchoRoutes(server interface{}) {
	// Importações específicas do Echo
	import (
		"net/http"
		"github.com/labstack/echo/v4"
		"github.com/fsvxavier/nexs-lib/httpservers/echo"
		pageecho "github.com/fsvxavier/nexs-lib/paginate/echo"
	)

	// Converter para o tipo específico
	echoServer := server.(*echo.EchoServer)
	e := echoServer.Echo()

	// Configurar rotas
	e.GET("/api/resources", func(c echo.Context) error {
		ctx := context.Background()

		// Extrair parâmetros de paginação
		metadata, err := pageecho.Parse(ctx, c, "id", "name", "created_at")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		}

		// Usar o serviço comum
		result, err := h.service.ListResources(ctx, metadata)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, result)
	})
}

// Exemplo de implementação do serviço
type ExampleResourceService struct {
	// campos do serviço
}

// ListResources implementa a interface ResourceService
func (s *ExampleResourceService) ListResources(ctx context.Context, metadata *page.Metadata) (*page.Output, error) {
	// Implementação da lógica de negócios
	// Esta parte permanece a mesma independentemente do framework

	// Dados simulados
	resources := []map[string]interface{}{
		{"id": 1, "name": "Resource 1", "created_at": "2023-01-01"},
		{"id": 2, "name": "Resource 2", "created_at": "2023-01-02"},
		{"id": 3, "name": "Resource 3", "created_at": "2023-01-03"},
	}

	// Criar saída paginada
	return page.NewOutputWithTotal(ctx, resources, len(resources), metadata)
}

// Exemplo de uso
func ExampleHttpServers() {
	// Criar serviço
	service := &ExampleResourceService{}

	// Configurar e iniciar servidor
	ConfigureServer(service)
}
```
*/
