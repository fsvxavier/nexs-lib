# ULD (ULID/UUID Provider)

Este pacote fornece uma abstração para geração, manipulação e conversão entre identificadores únicos ULID e UUID.

## Características

- Suporte para ULID (Universally Unique Lexicographically Sortable Identifier)
- Suporte para UUID (Universally Unique Identifier)
- Conversão entre formatos ULID e UUID
- Extração de timestamps de identificadores
- Validação de identificadores
- Operações seguras para concorrência
- Interface uniforme para diferentes tipos de identificadores

## Instalação

```bash
go get github.com/fsvxavier/nexs-lib/uld
```

## Uso Básico

### Criação de Identificadores

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/fsvxavier/nexs-lib/uld"
    "github.com/fsvxavier/nexs-lib/uld/interfaces"
)

func main() {
    // Criação de ULID
    ulid, err := uld.NewULID()
    if err != nil {
        panic(err)
    }
    fmt.Printf("ULID: %s\n", ulid.Value)
    fmt.Printf("ULID como UUID: %s\n", ulid.UUIDString)
    fmt.Printf("ULID Timestamp: %s\n", ulid.Timestamp)
    
    // Criação de UUID
    uuid, err := uld.NewUUID()
    if err != nil {
        panic(err)
    }
    fmt.Printf("UUID: %s\n", uuid.Value)
    fmt.Printf("UUID Timestamp: %s\n", uuid.Timestamp)
    
    // Criação com timestamp específico
    customTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
    customUlid, err := uld.NewULIDWithTime(customTime)
    if err != nil {
        panic(err)
    }
    fmt.Printf("ULID com timestamp específico: %s\n", customUlid.Value)
}
```

### Parsing e Validação

```go
package main

import (
    "fmt"
    
    "github.com/fsvxavier/nexs-lib/uld"
)

func main() {
    // Exemplo de string ULID
    ulidStr := "01H2XJWSXBH7BZJF3PA6G8K3RH"
    
    // Parse de ULID
    parsedUlid, err := uld.ParseULID(ulidStr)
    if err != nil {
        panic(err)
    }
    fmt.Printf("ULID parseado: %s\n", parsedUlid.Value)
    fmt.Printf("Timestamp: %s\n", parsedUlid.Timestamp)
    
    // Validação de ULID
    if uld.IsValidULID(ulidStr) {
        fmt.Println("ULID válido")
    } else {
        fmt.Println("ULID inválido")
    }
    
    // Exemplo de string UUID
    uuidStr := "f81d4fae-7dec-11d0-a765-00a0c91e6bf6"
    
    // Parse de UUID
    parsedUuid, err := uld.ParseUUID(uuidStr)
    if err != nil {
        panic(err)
    }
    fmt.Printf("UUID parseado: %s\n", parsedUuid.Value)
    
    // Validação de UUID
    if uld.IsValidUUID(uuidStr) {
        fmt.Println("UUID válido")
    } else {
        fmt.Println("UUID inválido")
    }
}
```

### Conversão entre Formatos

```go
package main

import (
    "fmt"
    
    "github.com/fsvxavier/nexs-lib/uld"
)

func main() {
    // Cria um novo ULID
    ulid, _ := uld.NewULID()
    
    // Converte para UUID
    uuidStr, err := uld.ToUUID(ulid.Value)
    if err != nil {
        panic(err)
    }
    fmt.Printf("ULID como UUID: %s\n", uuidStr)
    
    // Converte de UUID para ULID
    ulidStr, err := uld.FromUUID(uuidStr)
    if err != nil {
        panic(err)
    }
    fmt.Printf("UUID como ULID: %s\n", ulidStr)
    
    // Converte para hex
    hexStr, err := uld.ToHex(ulid.Value)
    if err != nil {
        panic(err)
    }
    fmt.Printf("ULID como hex: %s\n", hexStr)
    
    // Converte de hex para ULID
    backToUlid, err := uld.FromHex(hexStr)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Hex como ULID: %s\n", backToUlid)
}
```

### Extração de Timestamp

```go
package main

import (
    "fmt"
    
    "github.com/fsvxavier/nexs-lib/uld"
)

func main() {
    // Cria um novo ULID
    ulid, _ := uld.NewULID()
    
    // Extrai o timestamp
    timestamp, err := uld.ExtractTimestampFromULID(ulid.Value)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Timestamp do ULID: %s\n", timestamp)
    
    // Cria um novo UUID
    uuid, _ := uld.NewUUID()
    
    // Extrai o timestamp
    timestamp, err = uld.ExtractTimestampFromUUID(uuid.Value)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Timestamp do UUID: %s\n", timestamp)
}
```

## Uso Avançado

### Uso com Factory

```go
package main

import (
    "fmt"
    
    "github.com/fsvxavier/nexs-lib/uld"
    "github.com/fsvxavier/nexs-lib/uld/factory"
    "github.com/fsvxavier/nexs-lib/uld/interfaces"
)

func main() {
    // Cria uma fábrica personalizada
    f := factory.NewFactory()
    
    // Configura a fábrica com os provedores padrão
    factory := uld.SetupProviders()
    
    // Obtém um provedor específico
    ulidProvider, err := factory.GetProvider(interfaces.ULIDType)
    if err != nil {
        panic(err)
    }
    
    // Usa o provedor diretamente
    id, err := ulidProvider.New()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("ULID gerado pelo provedor: %s\n", id.Value)
}
```

## Considerações de Desempenho

- Os ULIDs são lexicograficamente ordenáveis, o que é útil para índices de banco de dados
- UUIDs são mais amplamente adotados e reconhecidos
- Em termos de performance, ambos são eficientes para geração de IDs únicos
- Para ordenação baseada em timestamp, ULIDs são preferíveis

## Thread Safety

Todos os provedores e conversores são seguros para uso concorrente.
