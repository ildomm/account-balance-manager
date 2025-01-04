package test_helpers

import (
	"context"
	"fmt"
	"github.com/ildomm/account-balance-manager/entity"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
	"math/rand"
	"sync"
)

// DatabaseMock is a mock type for the Querier type
type DatabaseMock struct {
	mock.Mock
	lock sync.Mutex

	keys      map[string]map[string]interface{}
	gameCount int
}

// NewDatabaseMock creates a new instance of MockQuerier
func NewDatabaseMock() *DatabaseMock {
	mocked := &DatabaseMock{
		keys: make(map[string]map[string]interface{}),
	}

	mocked.keys["game_results"] = make(map[string]interface{})
	mocked.gameCount = int(0)
	mocked.keys["user_balance"] = make(map[string]interface{})

	return mocked
}

func (m *DatabaseMock) Close() {
	m.Called()
}

func (m *DatabaseMock) GameCount() int {
	return m.gameCount
}

func (m *DatabaseMock) WithTransaction(ctx context.Context, fn func(*sqlx.Tx) error) (err error) {
	m.Called(ctx, fn)

	txn := new(sqlx.Tx)
	err = fn(txn)

	return err
}

func (m *DatabaseMock) InsertGameResult(ctx context.Context, txn sqlx.Tx, gameResult entity.GameResult) (int, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	args := m.Called(ctx, txn, gameResult)
	if len(args) > 0 {
		return 0, args.Error(1)
	} else {

		id := rand.Intn(1000)
		gameResult.ID = id

		m.keys["game_results"][fmt.Sprint(id)] = gameResult
		m.gameCount++

		return id, nil
	}
}

func (m *DatabaseMock) SelectUser(ctx context.Context, userID int) (*entity.User, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	args := m.Called(ctx, userID)

	if len(args) > 0 {
		if arg := args.Get(0); arg != nil {
			return arg.(*entity.User), nil
		}

		if arg := args.Get(1); arg != nil {
			return nil, args.Error(1)
		}
	}

	for _, user := range m.keys["user_balance"] {
		if user.(entity.User).ID == userID {
			_user := user.(entity.User)
			return &_user, nil
		}
	}

	return nil, nil
}

func (m *DatabaseMock) TransactionIDExist(ctx context.Context, transactionId string) (bool, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	args := m.Called(ctx, transactionId)
	if len(args) > 0 {
		return args.Bool(0), args.Error(1)
	} else {
		return false, nil
	}
}

func (m *DatabaseMock) UpdateUserBalance(ctx context.Context, txn sqlx.Tx, userID int, balance float64) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	user := entity.User{
		ID:      userID,
		Balance: balance,
	}

	m.keys["user_balance"][fmt.Sprint(userID)] = user

	args := m.Called(ctx, txn, userID, balance)

	if len(args) > 0 {
		return args.Error(0)
	} else {
		return nil
	}
}
