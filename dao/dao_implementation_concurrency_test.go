package dao

import (
	"context"
	"github.com/google/uuid"
	"github.com/ildomm/account-balance-manager/entity"
	"github.com/ildomm/account-balance-manager/test_helpers"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
)

func TestCreateGameResultConcurrentOnMock(t *testing.T) {
	databaseMock := test_helpers.NewDatabaseMock()
	ctx := context.Background()

	instance := NewGameResultDAO(databaseMock)

	userID := 1
	gameStatus := entity.GameStatusWin
	amount := 100.0
	transactionSource := entity.TransactionSourceGame

	// Mock successful interactions
	databaseMock.On("TransactionIDExist", ctx, mock.Anything)
	databaseMock.On("SelectUser", ctx, userID)
	databaseMock.On("WithTransaction", mock.Anything, mock.AnythingOfType("func(*sqlx.Tx) error"))
	databaseMock.On("InsertGameResult", ctx, mock.Anything, mock.Anything)
	databaseMock.On("UpdateUserBalance", ctx, mock.Anything, userID, mock.Anything)

	// Give to the mock a user with a balance of 0
	databaseMock.WithTransaction(ctx, func(txn *sqlx.Tx) error {
		databaseMock.UpdateUserBalance(ctx, *txn, userID, 0)
		return nil
	})

	// Inject many game results
	toInjectTotalEntries := [1000]int{} //nolint:all
	totalInjected := len(toInjectTotalEntries)
	expectedBalance := amount * float64(totalInjected)

	wg := sync.WaitGroup{}
	for range toInjectTotalEntries {
		wg.Add(1)

		// A go routine for each game result
		go func() {
			defer wg.Done()
			_, err := instance.CreateGameResult(ctx, userID, gameStatus, amount, transactionSource, uuid.New().String())
			assert.NoError(t, err)
		}()
	}

	// Wait for all workers to complete processing
	wg.Wait()

	// Basic mockers expectations check
	databaseMock.AssertExpectations(t)

	// Count the game results
	assert.Equal(t, databaseMock.GameCount(), totalInjected)

	// Compare the use balance
	user, err := databaseMock.SelectUser(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, user.Balance)
}

func TestCreateGameResultConcurrentWinLostOnMock(t *testing.T) {
	databaseMock := test_helpers.NewDatabaseMock()
	ctx := context.Background()

	instance := NewGameResultDAO(databaseMock)

	// Mock parameters
	userID := 1
	amountPerIteration := 100.0
	transactionSource := entity.TransactionSourceGame

	// The user must finish with a balance as it started
	initialBalance := 10000.0
	finalBalance := initialBalance

	// Total to inject of game results
	// To perform 100 Wins and 100 Losses
	toInjectTotalEntries := [100]int{} //nolint:all
	totalInjected := len(toInjectTotalEntries)

	// Mock successful interactions
	databaseMock.On("TransactionIDExist", ctx, mock.Anything)
	databaseMock.On("SelectUser", ctx, userID)
	databaseMock.On("WithTransaction", mock.Anything, mock.AnythingOfType("func(*sqlx.Tx) error"))
	databaseMock.On("InsertGameResult", ctx, mock.Anything, mock.Anything)
	databaseMock.On("UpdateUserBalance", ctx, mock.Anything, userID, mock.Anything).Times(201) // ( toInjectTotalEntries * 2 ) + 1

	// Give to the mock a user with a balance of 1000
	databaseMock.WithTransaction(ctx, func(txn *sqlx.Tx) error {

		// Must start with some balance, unless the user will have a negative balance for the first
		// entity.GameStatusLost hit
		databaseMock.UpdateUserBalance(ctx, *txn, userID, initialBalance)
		return nil
	})

	wg := sync.WaitGroup{}
	for range toInjectTotalEntries {
		wg.Add(1)

		// A go routine for each game result
		go func() {
			defer wg.Done()

			_, err := instance.CreateGameResult(ctx, userID, entity.GameStatusWin, amountPerIteration, transactionSource, uuid.New().String())
			assert.NoError(t, err)
		}()
	}

	// Even if all LOST operations ran first, the balance will be 0, never negative
	for range toInjectTotalEntries {
		wg.Add(1)

		// A go routine for each game result
		go func() {
			defer wg.Done()
			_, err := instance.CreateGameResult(ctx, userID, entity.GameStatusLost, amountPerIteration, transactionSource, uuid.New().String())
			assert.NoError(t, err)
		}()
	}

	// Wait for all workers to complete processing
	wg.Wait()

	// Basic mockers expectations check
	databaseMock.AssertExpectations(t)

	// Count the game results
	assert.Equal(t, databaseMock.GameCount(), totalInjected*2)

	// Compare the use balance
	user, err := databaseMock.SelectUser(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, finalBalance, user.Balance)
}
