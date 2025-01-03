package database

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type Querier interface {
	Close()
	WithTransaction(ctx context.Context, fn func(*sqlx.Tx) error) (err error)

	// TODO: Add the rest of the methods
}
