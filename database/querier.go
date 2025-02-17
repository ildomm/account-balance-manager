package database

import (
	"context"

	"github.com/ildomm/account-balance-manager/entity"
	"github.com/jmoiron/sqlx"
)

type Querier interface {
	Close()
	WithTransaction(ctx context.Context, fn func(*sqlx.Tx) error) (err error)

	SelectUser(ctx context.Context, userID int) (*entity.User, error)
	TransactionIDExist(ctx context.Context, transactionID string) (bool, error)

	InsertGameResult(ctx context.Context, txn sqlx.Tx, gameResult entity.GameResult) (int, error)
	UpdateUserBalance(ctx context.Context, txn sqlx.Tx, userID int, balance float64) error
}
