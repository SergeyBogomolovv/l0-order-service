package trm

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Transaction interface {
	Commit() error
	Rollback() error
}

type txKey struct{}

func withTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func ExtractTx(ctx context.Context) *sqlx.Tx {
	tx, ok := ctx.Value(txKey{}).(*sqlx.Tx)
	if !ok {
		return nil
	}
	return tx
}

type Manager interface {
	BeginTx(ctx context.Context) (context.Context, Transaction, error)
	Do(ctx context.Context, callback func(ctx context.Context) error) (err error)
}

type txManager struct {
	db *sqlx.DB
}

func NewManager(db *sqlx.DB) Manager {
	return &txManager{
		db: db,
	}
}

func (t *txManager) BeginTx(ctx context.Context) (context.Context, Transaction, error) {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	return withTx(ctx, tx), tx, nil
}

func (t *txManager) Do(ctx context.Context, callback func(ctx context.Context) error) error {
	ctx, tx, err := t.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := callback(ctx); err != nil {
		return err
	}
	return tx.Commit()
}
