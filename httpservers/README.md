# HTTP Servers

Esta biblioteca fornece implementações de servidores HTTP para diferentes frameworks de forma padronizada e fácil de usar. Suportamos as seguintes implementações:

- Fiber - Um framework web rápido inspirado em Express
- FastHTTP - Um servidor HTTP de alta performance
- net/http - A implementação padrão do Go
- Gin - Um framework web com excelente desempenho
- Echo - Um framework web minimalista e de alto desempenho
- Atreugo - Um framework web de alto desempenho baseado no FastHTTP

## Características

- Interface comum para diferentes implementações
- Graceful shutdown
- Reutilização de conexões
- Middleware para logging, tracing e recuperação de pânico
- Endpoints de saúde padrão
- Suporte para métricas, pprof e swagger

## Instalação

```bash
go get github.com/fsvxavier/nexs-lib
```

## Uso Básico

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/fsvxavier/nexs-lib/httpservers"
    "github.com/fsvxavier/nexs-lib/httpservers/common"
)

func main() {
    // Criar um servidor com Fiber
    server, err := httpservers.NewServer(
        httpservers.ServerTypeFiber,
        common.WithPort("8080"),
        common.WithHost("0.0.0.0"),
        common.WithReadTimeout(10 * time.Second),
        common.WithWriteTimeout(10 * time.Second),
        common.WithIdleTimeout(30 * time.Second),
        common.WithMetrics(true),
        common.WithPprof(true),
        common.WithSwagger(true),
    )
    if err != nil {
        panic(err)
    }
    
    // Iniciar o servidor (este método bloqueia até ser interrompido)
    if err := server.Start(); err != nil {
        fmt.Printf("Erro ao iniciar o servidor: %v\n", err)
    }
}
```

## Exemplos para Cada Implementação

### Fiber

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/fsvxavier/nexs-lib/httpservers"
    "github.com/fsvxavier/nexs-lib/httpservers/common"
    fiberServer "github.com/fsvxavier/nexs-lib/httpservers/fiber"
)

func main() {
    // Criar um servidor Fiber
    server, _ := httpservers.NewServer(httpservers.ServerTypeFiber, common.WithPort("8080"))
    
    // Acessar a instância subjacente do Fiber
    fiberInstance := server.(*fiberServer.FiberServer).App()
    
    // Adicionar rotas personalizadas
    fiberInstance.Get("/api/users", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"message": "Lista de usuários"})
    })
    
    // Iniciar o servidor
    server.Start()
}
```

### FastHTTP

```go
package main

import (
    "github.com/valyala/fasthttp"
    "github.com/fsvxavier/nexs-lib/httpservers"
    "github.com/fsvxavier/nexs-lib/httpservers/common"
    fasthttpServer "github.com/fsvxavier/nexs-lib/httpservers/fasthttp"
)

func main() {
    // Criar um servidor FastHTTP
    server, _ := httpservers.NewServer(httpservers.ServerTypeFastHTTP, common.WithPort("8080"))
    
    // Criar um handler personalizado
    customHandler := func(ctx *fasthttp.RequestCtx) {
        path := string(ctx.Path())
        
        switch {
        case path == "/api/users":
            ctx.SetContentType("application/json")
            ctx.WriteString(`{"message":"Lista de usuários"}`)
        default:
            // Passar para o handler padrão
            ctx.SetStatusCode(fasthttp.StatusNotFound)
            ctx.SetContentType("application/json")
            ctx.WriteString(`{"error":"Rota não encontrada"}`)
        }
    }
    
    // Configurar o handler personalizado
    fasthttpServer := server.(*fasthttpServer.FastHTTPServer)
    fasthttpServer.SetHandler(customHandler)
    
    // Iniciar o servidor
    server.Start()
}
```

### net/http

```go
package main

import (
    "net/http"
    "encoding/json"
    
    "github.com/fsvxavier/nexs-lib/httpservers"
    "github.com/fsvxavier/nexs-lib/httpservers/common"
    nethttpServer "github.com/fsvxavier/nexs-lib/httpservers/nethttp"
)

func main() {
    // Criar um servidor net/http
    server, _ := httpservers.NewServer(httpservers.ServerTypeNetHTTP, common.WithPort("8080"))
    
    // Acessar o router subjacente
    router := server.(*nethttpServer.NetHTTPServer).Router()
    
    // Adicionar rotas personalizadas
    router.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"message": "Lista de usuários"})
    })
    
    // Iniciar o servidor
    server.Start()
}
```

### Gin

```go
package main

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/fsvxavier/nexs-lib/httpservers"
    "github.com/fsvxavier/nexs-lib/httpservers/common"
    ginServer "github.com/fsvxavier/nexs-lib/httpservers/gin"
)

func main() {
    // Criar um servidor Gin
    server, _ := httpservers.NewServer(httpservers.ServerTypeGin, common.WithPort("8080"))
    
    // Acessar o router subjacente
    router := server.(*ginServer.GinServer).Router()
    
    // Adicionar rotas personalizadas
    router.GET("/api/users", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "Lista de usuários",
        })
    })
    
    // Iniciar o servidor
    server.Start()
}
```

### Echo

```go
package main

import (
    "net/http"
    
    "github.com/labstack/echo/v4"
    "github.com/fsvxavier/nexs-lib/httpservers"
    "github.com/fsvxavier/nexs-lib/httpservers/common"
    echoServer "github.com/fsvxavier/nexs-lib/httpservers/echo"
)

func main() {
    // Criar um servidor Echo
    server, _ := httpservers.NewServer(httpservers.ServerTypeEcho, common.WithPort("8080"))
    
    // Acessar a instância subjacente do Echo
    e := server.(*echoServer.EchoServer).Echo()
    
    // Adicionar rotas personalizadas
    e.GET("/api/users", func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]interface{}{
            "message": "Lista de usuários",
        })
    })
    
    // Iniciar o servidor
    server.Start()
}
```

### Atreugo

```go
package main

import (
    "net/http"
    
    "github.com/savsgio/atreugo/v11"
    "github.com/fsvxavier/nexs-lib/httpservers"
    "github.com/fsvxavier/nexs-lib/httpservers/common"
    atreugoServer "github.com/fsvxavier/nexs-lib/httpservers/atreugo"
)

func main() {
    // Criar um servidor Atreugo
    server, _ := httpservers.NewServer(httpservers.ServerTypeAtreugo, common.WithPort("8080"))
    
    // Acessar a instância subjacente do Atreugo
    atreugoInstance := server.(*atreugoServer.AtreugoServer).Server()
    
    // Adicionar rotas personalizadas
    atreugoInstance.GET("/api/users", func(ctx *atreugo.RequestCtx) error {
        return ctx.JSONResponse(map[string]interface{}{
            "message": "Lista de usuários",
        }, http.StatusOK)
    })
    
    // Iniciar o servidor
    server.Start()
}
```

## Gerenciamento Avançado

### Graceful Shutdown Personalizado

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/fsvxavier/nexs-lib/httpservers"
    "github.com/fsvxavier/nexs-lib/httpservers/common"
)

func main() {
    server, _ := httpservers.NewServer(httpservers.ServerTypeFiber)
    
    // Iniciar o servidor em uma goroutine
    go func() {
        if err := server.Start(); err != nil {
            panic(err)
        }
    }()
    
    // Aguardar sinal de interrupção
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    // Criar um contexto com timeout para desligamento
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Desligar o servidor
    if err := server.Shutdown(ctx); err != nil {
        panic(err)
    }
}
```
