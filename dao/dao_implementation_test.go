package dao

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ildomm/account-balance-manager/entity"
	"github.com/ildomm/account-balance-manager/test_helpers"
)

func TestCreateGameResultOnSuccess(t *testing.T) {
	databaseMock := test_helpers.NewDatabaseMock()
	ctx := context.Background()

	// Mock data
	userID := 1
	gameStatus := entity.GameStatusWin
	initialBalance := 100.0
	amount := 100.0
	finalBalance := initialBalance + amount
	transactionSource := entity.TransactionSourceGame
	transactionID := "unique-transaction-id"

	instance := NewAccountDAO(databaseMock)
	databaseMock.On("WithTransaction", mock.Anything, mock.AnythingOfType("func(*sqlx.Tx) error"))
	databaseMock.On("UpdateUserBalance", ctx, mock.Anything, userID, mock.Anything)

	// Create a fake user by forcing a balance over the Mock
	databaseMock.WithTransaction(ctx, func(txn *sqlx.Tx) error {
		databaseMock.UpdateUserBalance(ctx, *txn, 1, initialBalance)

		return nil
	})

	// Mock successful interactions
	databaseMock.On("TransactionIDExist", ctx, transactionID)
	databaseMock.On("SelectUser", ctx, userID)
	databaseMock.On("InsertGameResult", ctx, mock.Anything, mock.Anything)

	_, err := instance.CreateGameResult(ctx, userID, gameStatus, amount, transactionSource, transactionID)

	assert.NoError(t, err, "CreateGameResult should not return an error")
	databaseMock.AssertExpectations(t)

	// Count the game results
	assert.Equal(t, databaseMock.GameCount(), 1, "There should be one game result")

	// Check final balance
	user, err := databaseMock.SelectUser(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, finalBalance, user.Balance)
}

func TestCreateGameResultOnTransactionIDExists(t *testing.T) {
	databaseMock := test_helpers.NewDatabaseMock()

	instance := NewAccountDAO(databaseMock)

	ctx := context.Background()
	userID := 1
	gameStatus := entity.GameStatusWin
	amount := 100.0
	transactionSource := entity.TransactionSourceGame
	transactionID := "existing-transaction-id"

	// Mock transaction ID already exists
	databaseMock.On("TransactionIDExist", ctx, transactionID).Return(true, nil)

	_, err := instance.CreateGameResult(ctx, userID, gameStatus, amount, transactionSource, transactionID)

	assert.EqualError(t, err, entity.ErrTransactionIdExists.Error(), "CreateGameResult should return ErrTransactionIdExists")
	databaseMock.AssertExpectations(t)
}

func TestCreateGameResultOnUserNotFound(t *testing.T) {
	databaseMock := test_helpers.NewDatabaseMock()

	instance := NewAccountDAO(databaseMock)

	ctx := context.Background()
	userID := 1
	gameStatus := entity.GameStatusWin
	amount := 100.0
	transactionSource := entity.TransactionSourceGame
	transactionID := "unique-transaction-id"

	// Mock user not found
	databaseMock.On("TransactionIDExist", ctx, transactionID).Return(false, nil)
	databaseMock.On("SelectUser", ctx, userID).Return(nil, entity.ErrUserNotFound)

	_, err := instance.CreateGameResult(ctx, userID, gameStatus, amount, transactionSource, transactionID)

	assert.EqualError(t, err, entity.ErrUserNotFound.Error(), "CreateGameResult should return ErrUserNotFound")
	databaseMock.AssertExpectations(t)
}

func TestCreateGameResultOnInsufficientBalance(t *testing.T) {
	databaseMock := test_helpers.NewDatabaseMock()

	instance := NewAccountDAO(databaseMock)

	ctx := context.Background()
	userID := 1
	gameStatus := entity.GameStatusLose // Assuming this triggers the balance check
	amount := 300.0                     // Assuming the user's balance is less than this amount
	transactionSource := entity.TransactionSourceGame
	transactionID := "unique-transaction-id"

	// Mock user with insufficient balance
	databaseMock.On("TransactionIDExist", ctx, transactionID).Return(false, nil)
	databaseMock.On("SelectUser", ctx, userID).Return(&entity.User{
		ID:      userID,
		Balance: 200.0,
	}, nil)

	_, err := instance.CreateGameResult(ctx, userID, gameStatus, amount, transactionSource, transactionID)

	assert.EqualError(t, err, entity.ErrUserNegativeBalance.Error(), "CreateGameResult should return ErrUserNegativeBalance")
	databaseMock.AssertExpectations(t)
}

func TestCreateGameResultOnDatabaseError(t *testing.T) {
	databaseMock := test_helpers.NewDatabaseMock()

	instance := NewAccountDAO(databaseMock)

	ctx := context.Background()
	userID := 1
	gameStatus := entity.GameStatusWin
	amount := 100.0
	transactionSource := entity.TransactionSourceGame
	transactionID := "unique-transaction-id"

	// Mock successful interactions except for InsertGameResult
	databaseMock.On("TransactionIDExist", ctx, transactionID).Return(false, nil)
	databaseMock.On("SelectUser", ctx, userID).Return(&entity.User{
		ID:      userID,
		Balance: 200.0,
	}, nil)
	databaseMock.On("WithTransaction", mock.Anything, mock.AnythingOfType("func(*sqlx.Tx) error"))
	databaseMock.On("InsertGameResult", ctx, mock.Anything, mock.Anything).Return(nil, errors.New("database error"))

	_, err := instance.CreateGameResult(ctx, userID, gameStatus, amount, transactionSource, transactionID)

	assert.EqualError(t, err, entity.ErrCreatingGameResult.Error(), "CreateGameResult should return ErrCreatingGameResult")
	databaseMock.AssertExpectations(t)
}

func TestCreateGameResultOnSuccessMultiple(t *testing.T) {
	databaseMock := test_helpers.NewDatabaseMock()
	ctx := context.Background()

	instance := NewAccountDAO(databaseMock)

	userID := 1
	gameStatus := entity.GameStatusWin
	amount := 100.0
	transactionSource := entity.TransactionSourceGame
	transactionID := "unique-transaction-id"

	// Mock successful interactions
	databaseMock.On("TransactionIDExist", ctx, mock.Anything)
	databaseMock.On("SelectUser", ctx, userID).Return() // no fake results
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

	for range toInjectTotalEntries {
		transactionID = uuid.New().String()
		_, err := instance.CreateGameResult(ctx, userID, gameStatus, amount, transactionSource, transactionID)
		assert.NoError(t, err)
	}

	// Basic mockers expectations check
	databaseMock.AssertExpectations(t)

	// Count the game results
	assert.Equal(t, databaseMock.GameCount(), totalInjected)

	// Compare the use balance
	user, err := databaseMock.SelectUser(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, user.Balance)
}

func TestRetrieveUserOnSuccess(t *testing.T) {
	databaseMock := test_helpers.NewDatabaseMock()
	ctx := context.Background()

	instance := NewAccountDAO(databaseMock)

	userID := 1
	user := &entity.User{
		ID:      userID,
		Balance: 100.0,
	}

	databaseMock.On("SelectUser", ctx, userID).Return(user, nil)

	retrievedUser, err := instance.RetrieveUser(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, user, retrievedUser)
	databaseMock.AssertExpectations(t)
}

func TestRetrieveUserOnUserNotFound(t *testing.T) {
	databaseMock := test_helpers.NewDatabaseMock()
	ctx := context.Background()

	instance := NewAccountDAO(databaseMock)

	userID := 1

	databaseMock.On("SelectUser", ctx, userID).Return(nil, entity.ErrUserNotFound)

	_, err := instance.RetrieveUser(ctx, userID)

	assert.EqualError(t, err, entity.ErrUserNotFound.Error())
	databaseMock.AssertExpectations(t)

}
func TestRetrieveUserOnDatabaseError(t *testing.T) {
	databaseMock := test_helpers.NewDatabaseMock()
	ctx := context.Background()

	instance := NewAccountDAO(databaseMock)

	userID := 1
	databaseError := errors.New("database error")

	databaseMock.On("SelectUser", ctx, userID).Return(nil, databaseError)

	_, err := instance.RetrieveUser(ctx, userID)

	assert.EqualError(t, err, databaseError.Error())
	databaseMock.AssertExpectations(t)
}
