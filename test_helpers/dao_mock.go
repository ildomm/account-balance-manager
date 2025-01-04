package test_helpers

import (
	"context"
	"github.com/ildomm/account-balance-manager/entity"
	"github.com/stretchr/testify/mock"
)

// daoMock is a mock type for the DAO type
type DAOMock struct {
	mock.Mock
}

// NewDAOMock creates a new instance of daoMock
func NewDAOMock() *DAOMock {
	return &DAOMock{}
}

func (m *DAOMock) CreateGameResult(
	ctx context.Context,
	userID int,
	gameStatus entity.GameStatus,
	amount float64,
	transactionSource entity.TransactionSource,
	transactionID string) (*entity.GameResult, error) {

	args := m.Called(ctx, userID, gameStatus, amount, transactionSource, transactionID)

	if len(args) > 0 {
		if arg := args.Get(0); arg != nil {
			return arg.(*entity.GameResult), nil
		}
		if args.Get(1) != nil {
			return nil, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

func (m *DAOMock) RetrieveUser(ctx context.Context, userID int) (*entity.User, error) {
	args := m.Called(ctx, userID)

	if len(args) > 0 {
		if arg := args.Get(0); arg != nil {
			return arg.(*entity.User), nil
		}
		if args.Get(1) != nil {
			return nil, args.Error(1)
		}
	}
	return nil, args.Error(1)
}
