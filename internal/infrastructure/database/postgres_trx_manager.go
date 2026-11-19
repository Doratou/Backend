package database

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

type TrxManager interface {
	WithTransaction(ctx context.Context, fn func(trxCtx context.Context) error) error
}

type gormTrxManager struct {
	db      Database
	options *sql.TxOptions
}

func NewTrxManager(db Database) TrxManager {
	return &gormTrxManager{db: db}
}

func NewReadSnapshotTrxManager(db Database) TrxManager {
	return &gormTrxManager{db: db, options: &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true}}
}

func (m *gormTrxManager) WithTransaction(ctx context.Context, fn func(trxCtx context.Context) error) error {
	return m.db.GetGormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDatabase := m.db.WithTx(tx)
		trxCtx := InjectTrx(ctx, txDatabase)
		return fn(trxCtx)
	}, m.options)
}
