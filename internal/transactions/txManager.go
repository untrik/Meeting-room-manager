package transactions

import (
	"context"
	"database/sql"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type sQLTxManager struct {
	db *sql.DB
}

func NewSQLTxManager(db *sql.DB) *sQLTxManager {
	return &sQLTxManager{db: db}
}

type txKey struct{}

func (sqlTx *sQLTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := sqlTx.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
func TxFromContext(ctx context.Context) (DBTX, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if !ok {
		return nil, false
	}
	return tx, true
}
