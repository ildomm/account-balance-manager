package dao

import (
	"context"
	"github.com/ildomm/account-balance-manager/entity"
)

type DAO interface {
	CreateGameResult(ctx context.Context, userID int, gameStatus entity.GameStatus, amount float64, transactionSource entity.TransactionSource, transactionID string) (*entity.GameResult, error)
	RetrieveUser(ctx context.Context, userID int) (*entity.User, error)
}
