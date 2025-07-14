//go:build unit

// Package mocks contains generated mocks for the PGX provider
package mocks

//go:generate mockgen -source=../../interfaces.go -destination=mock_interfaces.go -package=mocks IProvider,IPool,IConn,ITransaction,IBatch,IBatchResults,IRows,IRow
//go:generate mockgen -source=../provider.go -destination=mock_provider_impl.go -package=mocks Provider
//go:generate mockgen -source=../pool.go -destination=mock_pool_impl.go -package=mocks Pool
//go:generate mockgen -source=../conn.go -destination=mock_conn_impl.go -package=mocks Conn
//go:generate mockgen -source=../transaction.go -destination=mock_transaction_impl.go -package=mocks Transaction
//go:generate mockgen -source=../rows.go -destination=mock_rows_impl.go -package=mocks Rows,Row,Batch,BatchResults
