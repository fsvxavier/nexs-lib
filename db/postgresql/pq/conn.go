package pq

import (
	"context"
	"database/sql"
	"errors"
	"reflect"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
)

// Conn implementa a interface common.IConn usando database/sql e lib/pq
type Conn struct {
	conn               *sql.Conn
	db                 *sql.DB
	multiTenantEnabled bool
	config             *common.Config
}

// NewConn cria uma nova conexão direta (sem pool) com o PostgreSQL usando lib/pq
func NewConn(ctx context.Context, config *common.Config) (common.IConn, error) {
	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		return nil, err
	}

	// Adquire uma conexão
	conn, err := db.Conn(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Configura o fuso horário para UTC
	_, err = conn.ExecContext(ctx, "SET timezone TO 'UTC'")
	if err != nil {
		conn.Close()
		db.Close()
		return nil, err
	}

	pqConn := &Conn{
		conn:               conn,
		db:                 db,
		multiTenantEnabled: config.MultiTenantEnabled,
		config:             config,
	}

	// Se multi-tenant está habilitado, configura o contexto do tenant
	if config.MultiTenantEnabled {
		if err := pqConn.setupTenantContext(ctx); err != nil {
			conn.Close()
			db.Close()
			return nil, err
		}
	}

	return pqConn, nil
}

// QueryOne executa uma consulta e digitaliza uma única linha no destino
func (c *Conn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := c.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanOne(rows, dst)
}

// QueryAll executa uma consulta e digitaliza todas as linhas no destino
func (c *Conn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := c.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanAll(rows, dst)
}

// QueryCount executa uma consulta e retorna uma contagem
func (c *Conn) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	var count int

	row, err := c.QueryRow(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	err = row.Scan(&count)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// Query executa uma consulta e retorna as linhas resultantes
func (c *Conn) Query(ctx context.Context, query string, args ...interface{}) (common.IRows, error) {
	rows, err := c.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

// QueryRow executa uma consulta e retorna uma única linha
func (c *Conn) QueryRow(ctx context.Context, query string, args ...interface{}) (common.IRow, error) {
	row := c.conn.QueryRowContext(ctx, query, args...)
	return &Row{row: row}, nil
}

// Exec executa um comando (sem retorno de linhas)
func (c *Conn) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := c.conn.ExecContext(ctx, query, args...)
	return err
}

// SendBatch envia um lote de comandos para execução
func (c *Conn) SendBatch(ctx context.Context, batch common.IBatch) (common.IBatchResults, error) {
	pqBatch, ok := batch.(*Batch)
	if !ok {
		return nil, common.ErrInvalidOperation
	}

	tx, err := c.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &BatchResults{
		tx:    tx,
		ctx:   ctx,
		batch: pqBatch,
	}, nil
}

// Ping verifica a conectividade com o banco de dados
func (c *Conn) Ping(ctx context.Context) error {
	return c.conn.PingContext(ctx)
}

// Close fecha a conexão
func (c *Conn) Close(ctx context.Context) error {
	if c.conn != nil {
		err := c.conn.Close()
		if c.db != nil {
			c.db.Close()
		}
		return err
	}
	return nil
}

// BeginTransaction inicia uma transação
func (c *Conn) BeginTransaction(ctx context.Context) (common.ITransaction, error) {
	return c.BeginTransactionWithOptions(ctx, nil)
}

// BeginTransactionWithOptions inicia uma transação com opções específicas
func (c *Conn) BeginTransactionWithOptions(ctx context.Context, opts *common.TxOptions) (common.ITransaction, error) {
	var options *sql.TxOptions

	if opts != nil {
		options = &sql.TxOptions{
			ReadOnly: opts.ReadOnly,
		}

		// Mapeia os níveis de isolamento
		switch opts.IsoLevel {
		case common.ReadCommitted:
			options.Isolation = sql.LevelReadCommitted
		case common.RepeatableRead:
			options.Isolation = sql.LevelRepeatableRead
		case common.Serializable:
			options.Isolation = sql.LevelSerializable
		default:
			options.Isolation = sql.LevelDefault
		}
	}

	tx, err := c.conn.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}

	// Se multi-tenant está habilitado, configura o contexto do tenant na transação
	if c.multiTenantEnabled {
		tenantID := getTenantIDFromContext(ctx)
		if tenantID != "" {
			_, err = tx.ExecContext(ctx, "SET app.tenant_id = $1", tenantID)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	return &Transaction{
		tx:                 tx,
		config:             c.config,
		multiTenantEnabled: c.multiTenantEnabled,
	}, nil
}

// setupTenantContext configura o contexto do tenant para multi-tenancy
func (c *Conn) setupTenantContext(ctx context.Context) error {
	// Implementação da configuração do tenant
	// Este é um esboço - a implementação real dependeria da sua abordagem de multi-tenancy
	tenantID := getTenantIDFromContext(ctx)
	if tenantID == "" {
		return nil // Nenhum tenant especificado, não é necessário configurar
	}

	// Exemplo: definir um parâmetro de sessão para o tenant atual
	query := "SET app.tenant_id = $1"
	return c.Exec(ctx, query, tenantID)
}

// getTenantIDFromContext extrai o ID do tenant do contexto
func getTenantIDFromContext(ctx context.Context) string {
	// Implementação real dependeria da sua estrutura de contexto
	// Exemplo simples:
	type tenantKey struct{}
	if tenant, ok := ctx.Value(tenantKey{}).(string); ok {
		return tenant
	}
	return ""
}

// scanOne digitaliza uma única linha no destino
func scanOne(rows *sql.Rows, dst interface{}) error {
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return common.ErrNoRows
	}

	err := scanRow(rows, dst)
	if err != nil {
		return err
	}

	// Verifica se há mais linhas, o que seria um erro para scanOne
	if rows.Next() {
		return errors.New("mais de uma linha retornada para scanOne")
	}

	return rows.Err()
}

// scanAll digitaliza todas as linhas no destino
func scanAll(rows *sql.Rows, dst interface{}) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return errors.New("destino deve ser um ponteiro não nulo")
	}

	sliceVal := reflect.Indirect(dstVal)
	if sliceVal.Kind() != reflect.Slice {
		return errors.New("destino deve ser um ponteiro para um slice")
	}

	elemType := sliceVal.Type().Elem()
	isPtr := elemType.Kind() == reflect.Ptr

	var elemTypeForNew reflect.Type
	if isPtr {
		elemTypeForNew = elemType.Elem()
	} else {
		elemTypeForNew = elemType
	}

	for rows.Next() {
		newElem := reflect.New(elemTypeForNew)

		if err := scanRow(rows, newElem.Interface()); err != nil {
			return err
		}

		if !isPtr {
			newElem = newElem.Elem()
		}

		sliceVal = reflect.Append(sliceVal, newElem)
	}

	reflect.Indirect(dstVal).Set(sliceVal)
	return rows.Err()
}

// scanRow digitaliza uma linha em um destino
func scanRow(rows *sql.Rows, dst interface{}) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return errors.New("destino deve ser um ponteiro não nulo")
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// Se for um struct, mapeamos os campos pelos nomes das colunas
	dstElem := dstVal.Elem()
	if dstElem.Kind() == reflect.Struct {
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(interface{})
		}

		if err := rows.Scan(values...); err != nil {
			return err
		}

		// Mapeamos valores para campos do struct
		for i, col := range columns {
			fieldVal := findField(dstElem, col)
			if fieldVal.IsValid() && fieldVal.CanSet() {
				setValue(fieldVal, *(values[i].(*interface{})))
			}
		}

		return nil
	}

	// Caso contrário, usamos scan diretamente
	return rows.Scan(dst)
}

// findField encontra um campo em um struct pelo nome da coluna
func findField(v reflect.Value, colName string) reflect.Value {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")
		if tag == colName || field.Name == colName {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}

// setValue define o valor de um campo
func setValue(field reflect.Value, value interface{}) {
	if value == nil {
		return
	}

	valueType := reflect.TypeOf(value)
	fieldType := field.Type()

	// Se os tipos são compatíveis, definimos diretamente
	if valueType.AssignableTo(fieldType) {
		field.Set(reflect.ValueOf(value))
		return
	}

	// Tentamos conversões básicas
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := value.(type) {
		case int64:
			field.SetInt(v)
		case int:
			field.SetInt(int64(v))
		case float64:
			field.SetInt(int64(v))
		case string:
			// Tenta converter string para int
			// Na implementação real, usaria strconv.ParseInt
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch v := value.(type) {
		case uint64:
			field.SetUint(v)
		case int:
			field.SetUint(uint64(v))
		case float64:
			field.SetUint(uint64(v))
		}
	case reflect.Float32, reflect.Float64:
		switch v := value.(type) {
		case float64:
			field.SetFloat(v)
		case int:
			field.SetFloat(float64(v))
		}
	case reflect.String:
		switch v := value.(type) {
		case string:
			field.SetString(v)
		case []byte:
			field.SetString(string(v))
		}
	case reflect.Bool:
		switch v := value.(type) {
		case bool:
			field.SetBool(v)
		case int:
			field.SetBool(v != 0)
		}
	}
}
