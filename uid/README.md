# UID Module

O módulo UID oferece uma biblioteca abrangente para geração e manipulação de identificadores únicos em Go, suportando múltiplos tipos de UID incluindo ULID e variantes UUID com uma interface unificada.

## Características

### Tipos de UID Suportados
- **ULID** (Universally Unique Lexicographically Sortable Identifier)
- **UUID v1** (baseado em timestamp e endereço MAC)
- **UUID v4** (gerado aleatoriamente)
- **UUID v6** (versão ordenada do UUID v1)
- **UUID v7** (baseado em timestamp Unix)

### Funcionalidades Principais
- ✅ Geração thread-safe de UIDs
- ✅ Suporte a marshal/unmarshal (Text, Binary, JSON)
- ✅ Validação e parsing automático
- ✅ Detecção automática de formato
- ✅ Cache de providers para performance
- ✅ Suporte a timestamp personalizado
- ✅ Arquitetura hexagonal com injeção de dependência
- ✅ Configuração flexível e validação

## Instalação

```bash
go get github.com/fsvxavier/nexs-lib/uid
```

## Uso Básico

### Criação de Manager com Configuração Padrão

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/fsvxavier/nexs-lib/uid"
    "github.com/fsvxavier/nexs-lib/uid/interfaces"
)

func main() {
    // Criar manager com configuração padrão
    manager, err := uid.NewDefaultUIDManager()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Gerar ULID
    ulid, err := manager.Generate(ctx, interfaces.UIDTypeULID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("ULID: %s\n", ulid.Canonical)
    fmt.Printf("Timestamp: %v\n", ulid.Timestamp)
    fmt.Printf("Hex: %s\n", ulid.Hex)
}
```

### Geração de Diferentes Tipos de UID

```go
// Gerar UUID v4
uuid4, err := manager.Generate(ctx, interfaces.UIDTypeUUIDV4)
if err != nil {
    log.Fatal(err)
}

// Gerar UUID v7 com timestamp específico
timestamp := time.Now()
uuid7, err := manager.GenerateWithTimestamp(ctx, interfaces.UIDTypeUUIDV7, timestamp)
if err != nil {
    log.Fatal(err)
}

// Gerar UID usando tipo padrão
defaultUID, err := manager.GenerateDefault(ctx)
if err != nil {
    log.Fatal(err)
}
```

### Parsing e Validação

```go
// Parse automático com detecção de formato
input := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
uid, err := manager.Parse(ctx, input)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Tipo detectado: %s\n", uid.Type)

// Validação
err = manager.Validate(ctx, input)
if err != nil {
    log.Printf("UID inválido: %v", err)
}
```

## Configuração Avançada

### Configuração Personalizada da Factory

```go
package main

import (
    "github.com/fsvxavier/nexs-lib/uid"
    "github.com/fsvxavier/nexs-lib/uid/config"
    "github.com/fsvxavier/nexs-lib/uid/interfaces"
)

func main() {
    // Configuração personalizada
    cfg := &config.FactoryConfig{
        DefaultType:    interfaces.UIDTypeULID,
        CacheTimeout:   time.Minute * 30,
        MaxCacheSize:   1000,
        EnableMetrics:  true,
    }

    // Criar factory com configuração personalizada
    factory, err := uid.NewFactory(cfg)
    if err != nil {
        log.Fatal(err)
    }

    manager := uid.NewUIDManager(factory)
}
```

### Configuração de Providers Específicos

```go
// Configuração ULID personalizada
ulidConfig := &config.ULIDConfig{
    Provider: config.ProviderConfig{
        Type:      interfaces.UIDTypeULID,
        Name:      "custom-ulid",
        CacheSize: 100,
    },
    EntropySize: 10, // bytes de entropia
}

// Configuração UUID personalizada
uuidConfig := &config.UUIDConfig{
    Provider: config.ProviderConfig{
        Type:      interfaces.UIDTypeUUIDV4,
        Name:      "custom-uuid",
        CacheSize: 50,
    },
    Version:       4,
    NodeID:        []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
    ClockSequence: 12345,
}
```

## Serialização e Marshal/Unmarshal

### Text Marshal/Unmarshal

```go
// Marshal para texto
textData, err := uid.MarshalText()
if err != nil {
    log.Fatal(err)
}

// Unmarshal do texto
newUID := &interfaces.UIDData{}
err = newUID.UnmarshalText(textData)
if err != nil {
    log.Fatal(err)
}
```

### JSON Marshal/Unmarshal

```go
// Marshal para JSON
jsonData, err := uid.MarshalJSON()
if err != nil {
    log.Fatal(err)
}

// Unmarshal do JSON
newUID := &interfaces.UIDData{}
err = newUID.UnmarshalJSON(jsonData)
if err != nil {
    log.Fatal(err)
}
```

### Binary Marshal/Unmarshal

```go
// Marshal para binário
binaryData, err := uid.MarshalBinary()
if err != nil {
    log.Fatal(err)
}

// Unmarshal do binário
newUID := &interfaces.UIDData{}
err = newUID.UnmarshalBinary(binaryData)
if err != nil {
    log.Fatal(err)
}
```

## Estrutura de Dados

### UIDData

```go
type UIDData struct {
    Raw       string     // Formato original
    Canonical string     // Formato canônico
    Bytes     []byte     // Representação em bytes
    Hex       string     // Representação hexadecimal
    Type      UIDType    // Tipo do UID
    Timestamp *time.Time // Timestamp extraído (se aplicável)
    Version   *int       // Versão UUID (se aplicável)
    Variant   *string    // Variante UUID (se aplicável)
}
```

## Operações Concorrentes

O módulo é totalmente thread-safe e otimizado para operações concorrentes:

```go
func ExemploConcorrencia() {
    manager, _ := uid.NewDefaultUIDManager()
    ctx := context.Background()

    // Gerar UIDs concorrentemente
    const numGoroutines = 100
    results := make(chan *interfaces.UIDData, numGoroutines)

    for i := 0; i < numGoroutines; i++ {
        go func() {
            uid, err := manager.Generate(ctx, interfaces.UIDTypeULID)
            if err == nil {
                results <- uid
            }
        }()
    }

    // Coletar resultados
    uids := make([]string, 0, numGoroutines)
    for i := 0; i < numGoroutines; i++ {
        uid := <-results
        uids = append(uids, uid.Canonical)
    }
}
```

## Tratamento de Erros

O módulo define tipos de erro específicos para diferentes cenários:

```go
// Erro de validação
var validationErr *internal.ValidationError
if errors.As(err, &validationErr) {
    fmt.Printf("Erro de validação: %s\n", validationErr.Error())
}

// Erro de parsing
var parseErr *internal.ParseError
if errors.As(err, &parseErr) {
    fmt.Printf("Erro de parsing: %s\n", parseErr.Error())
}

// Erro de marshal/unmarshal
var marshalErr *interfaces.MarshalError
if errors.As(err, &marshalErr) {
    fmt.Printf("Erro de marshal: %s\n", marshalErr.Error())
}
```

## Exemplos Práticos

### Sistema de Identificação de Entidades

```go
type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func CreateUser(manager *uid.UIDManager, name string) (*User, error) {
    ctx := context.Background()
    
    // Gerar ID único para o usuário
    userID, err := manager.Generate(ctx, interfaces.UIDTypeULID)
    if err != nil {
        return nil, err
    }

    return &User{
        ID:   userID.Canonical,
        Name: name,
    }, nil
}
```

### Log Correlation ID

```go
func HandleRequest(manager *uid.UIDManager, w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Gerar correlation ID
    correlationID, err := manager.Generate(ctx, interfaces.UIDTypeULID)
    if err != nil {
        http.Error(w, "Internal error", 500)
        return
    }

    // Adicionar ao contexto e headers
    ctx = context.WithValue(ctx, "correlation_id", correlationID.Canonical)
    w.Header().Set("X-Correlation-ID", correlationID.Canonical)
    
    // Processar request...
}
```

## Performance

- Operações de geração: ~100ns por UID
- Cache de providers para reutilização eficiente
- Pool de objetos para reduzir garbage collection
- Operações lock-free onde possível

## Dependências

- `github.com/oklog/ulid/v2` - Para geração de ULID
- `github.com/google/uuid` - Para geração de UUID
- `github.com/stretchr/testify` - Para testes (apenas dev)

## Compatibilidade

- Go 1.21+
- Thread-safe
- Compatível com JSON, XML, YAML marshaling
- Suporta context.Context para cancelamento

## Licença

Este módulo faz parte da biblioteca nexs-lib e segue a mesma licença do projeto principal.
