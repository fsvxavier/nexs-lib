package pgxprovider

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// ReflectionCache cache de informações de reflection para performance
type ReflectionCache struct {
	mu    sync.RWMutex
	cache map[reflect.Type]*StructInfo
}

// StructInfo informações sobre um struct para mapeamento
type StructInfo struct {
	Fields    []FieldInfo
	FieldsMap map[string]int // nome -> índice no slice Fields
}

// FieldInfo informações sobre um campo
type FieldInfo struct {
	Name       string
	Index      int
	Type       reflect.Type
	DBColumn   string
	IsPointer  bool
	IsNullable bool
	Converter  ValueConverter
}

// ValueConverter interface para conversão de valores
type ValueConverter interface {
	Convert(src interface{}) (interface{}, error)
}

// DefaultConverter conversor padrão
type DefaultConverter struct{}

func (dc *DefaultConverter) Convert(src interface{}) (interface{}, error) {
	return src, nil
}

// TimeConverter conversor para time.Time
type TimeConverter struct{}

func (tc *TimeConverter) Convert(src interface{}) (interface{}, error) {
	if src == nil {
		return nil, nil
	}

	switch v := src.(type) {
	case time.Time:
		return v, nil
	case string:
		return time.Parse(time.RFC3339, v)
	default:
		return nil, fmt.Errorf("cannot convert %T to time.Time", src)
	}
}

// Cache global de reflection
var reflectionCache = &ReflectionCache{
	cache: make(map[reflect.Type]*StructInfo),
}

// getStructInfo obtém informações sobre um struct (com cache)
func (rc *ReflectionCache) getStructInfo(structType reflect.Type) (*StructInfo, error) {
	rc.mu.RLock()
	if info, exists := rc.cache[structType]; exists {
		rc.mu.RUnlock()
		return info, nil
	}
	rc.mu.RUnlock()

	// Não está no cache, vamos criar
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Double-check após obter write lock
	if info, exists := rc.cache[structType]; exists {
		return info, nil
	}

	info, err := rc.analyzeStruct(structType)
	if err != nil {
		return nil, err
	}

	rc.cache[structType] = info
	return info, nil
}

// analyzeStruct analisa um struct e cria StructInfo
func (rc *ReflectionCache) analyzeStruct(structType reflect.Type) (*StructInfo, error) {
	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v", structType.Kind())
	}

	info := &StructInfo{
		Fields:    make([]FieldInfo, 0),
		FieldsMap: make(map[string]int),
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Pular campos não exportados
		if !field.IsExported() {
			continue
		}

		fieldInfo := FieldInfo{
			Name:      field.Name,
			Index:     i,
			Type:      field.Type,
			IsPointer: field.Type.Kind() == reflect.Ptr,
			Converter: &DefaultConverter{},
		}

		// Extrair tag db
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			// Se não tem tag db, usar nome do campo em snake_case
			fieldInfo.DBColumn = toSnakeCase(field.Name)
		} else if dbTag == "-" {
			// Tag "-" significa ignorar este campo
			continue
		} else {
			// Usar tag como nome da coluna
			fieldInfo.DBColumn = dbTag
		}

		// Verificar se é nullable
		if fieldInfo.IsPointer {
			fieldInfo.IsNullable = true
		}

		// Configurar conversor específico por tipo
		baseType := fieldInfo.Type
		if baseType.Kind() == reflect.Ptr {
			baseType = baseType.Elem()
		}

		switch baseType {
		case reflect.TypeOf(time.Time{}):
			fieldInfo.Converter = &TimeConverter{}
		}

		info.Fields = append(info.Fields, fieldInfo)
		info.FieldsMap[fieldInfo.DBColumn] = len(info.Fields) - 1
	}

	return info, nil
}

// toSnakeCase converte CamelCase para snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// queryAllWithReflection implementa mapeamento automático usando reflection
func (c *Conn) queryAllWithReflection(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	// Validar destino
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	// Obter tipo do elemento da slice
	elementType := destValue.Type().Elem()
	isPointer := elementType.Kind() == reflect.Ptr
	if isPointer {
		elementType = elementType.Elem()
	}

	if elementType.Kind() != reflect.Struct {
		return fmt.Errorf("slice element must be a struct")
	}

	// Obter informações do struct
	structInfo, err := reflectionCache.getStructInfo(elementType)
	if err != nil {
		return fmt.Errorf("failed to analyze struct: %w", err)
	}

	// Executar query
	rows, err := c.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Obter descrições das colunas
	fieldDescs := rows.FieldDescriptions()
	columnIndexes := make([]int, len(fieldDescs))

	// Mapear colunas para campos do struct
	for i, desc := range fieldDescs {
		columnName := desc.Name()
		if fieldIndex, exists := structInfo.FieldsMap[columnName]; exists {
			columnIndexes[i] = fieldIndex
		} else {
			columnIndexes[i] = -1 // Coluna não mapeada
		}
	}

	// Processar todas as linhas
	results := reflect.MakeSlice(destValue.Type(), 0, 0)

	for rows.Next() {
		// Criar nova instância do struct
		var elementValue reflect.Value
		if isPointer {
			elementValue = reflect.New(elementType)
		} else {
			elementValue = reflect.New(elementType).Elem()
		}

		// Criar slice para valores das colunas
		scanTargets := make([]interface{}, len(fieldDescs))

		for i, fieldIndex := range columnIndexes {
			if fieldIndex == -1 {
				// Coluna não mapeada, usar interface{} genérico
				var dummy interface{}
				scanTargets[i] = &dummy
				continue
			}

			field := structInfo.Fields[fieldIndex]

			if field.IsNullable {
				// Para campos nullable, usar ponteiro
				scanTargets[i] = reflect.New(field.Type.Elem()).Interface()
			} else {
				// Para campos não nullable, usar valor direto
				scanTargets[i] = reflect.New(field.Type).Interface()
			}
		} // Fazer scan da linha
		if err := rows.Scan(scanTargets...); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		// Converter e atribuir valores
		for i, fieldIndex := range columnIndexes {
			if fieldIndex == -1 {
				continue
			}

			field := structInfo.Fields[fieldIndex]
			scannedValue := reflect.ValueOf(scanTargets[i]).Elem()

			var target reflect.Value
			if isPointer {
				target = elementValue.Elem().Field(field.Index)
			} else {
				target = elementValue.Field(field.Index)
			}

			// Converter valor se necessário
			convertedValue, err := field.Converter.Convert(scannedValue.Interface())
			if err != nil {
				return fmt.Errorf("conversion failed for field %s: %w", field.Name, err)
			}

			// Atribuir valor
			if convertedValue == nil && field.IsNullable {
				// Valor null para campo nullable
				target.Set(reflect.Zero(field.Type))
			} else if convertedValue != nil {
				convertedReflectValue := reflect.ValueOf(convertedValue)
				if field.IsNullable {
					// Para campos nullable, criar ponteiro
					ptrValue := reflect.New(field.Type.Elem())
					ptrValue.Elem().Set(convertedReflectValue)
					target.Set(ptrValue)
				} else {
					target.Set(convertedReflectValue)
				}
			}
		}

		// Adicionar elemento ao resultado
		if isPointer {
			results = reflect.Append(results, elementValue)
		} else {
			results = reflect.Append(results, elementValue)
		}
	}

	// Verificar erros do rows
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}

	// Atribuir resultado ao destino
	destValue.Set(results)

	return nil
}

// QueryOneWithReflection implementa mapeamento automático para uma única linha
func (c *Conn) QueryOneWithReflection(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	// Criar slice temporário para usar QueryAll
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	destType := destValue.Type()
	sliceType := reflect.SliceOf(destType)
	tempSlice := reflect.MakeSlice(sliceType, 0, 1)
	tempSlicePtr := reflect.New(sliceType)
	tempSlicePtr.Elem().Set(tempSlice)

	// Usar QueryAll para obter resultado
	err := c.QueryAll(ctx, tempSlicePtr.Interface(), query, args...)
	if err != nil {
		return err
	}

	// Verificar se obtivemos resultado
	resultSlice := tempSlicePtr.Elem()
	if resultSlice.Len() == 0 {
		return pgx.ErrNoRows
	}

	// Atribuir primeiro elemento ao destino
	destValue.Elem().Set(resultSlice.Index(0).Elem())

	return nil
}

// ReflectionStats estatísticas de uso da reflection
type ReflectionStats struct {
	CacheHits        int64
	CacheMisses      int64
	StructsAnalyzed  int64
	ConversionsCount int64
}

// GetReflectionStats retorna estatísticas de uso
func (rc *ReflectionCache) GetReflectionStats() ReflectionStats {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	return ReflectionStats{
		StructsAnalyzed: int64(len(rc.cache)),
		// Outras estatísticas podem ser adicionadas conforme necessário
	}
}

// ClearReflectionCache limpa o cache de reflection
func (rc *ReflectionCache) ClearReflectionCache() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.cache = make(map[reflect.Type]*StructInfo)
}
