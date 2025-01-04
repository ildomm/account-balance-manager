package dao

import (
	"context"
	"fmt"
	"github.com/ildomm/account-balance-manager/database"
	"github.com/ildomm/account-balance-manager/entity"
	"github.com/jmoiron/sqlx"
	"log"
	"sync"
	"time"
)

type accountDAO struct {
	querier database.Querier
	lock    sync.Mutex
}

// NewAccountDAO creates a new game result DAO
func NewAccountDAO(querier database.Querier) *accountDAO {
	return &accountDAO{querier: querier}
}

// CreateGameResult creates a new game result
// It validates the transaction and updates the user balance
// It returns the created game result
// It returns an error if the transaction is invalid or if there is an error creating the game result
func (dm *accountDAO) CreateGameResult(ctx context.Context, userID int, gameStatus entity.GameStatus, amount float64, transactionSource entity.TransactionSource, transactionID string) (*entity.GameResult, error) {
	dm.lock.Lock()
	defer dm.lock.Unlock()

	// Check the transaction and its related user
	balance := 0.0
	if user, err := dm.validateTransaction(ctx, userID, gameStatus, amount, transactionID); err != nil {
		return nil, err
	} else {
		balance = dm.calculateNewBalance(user.Balance, gameStatus, amount)
	}

	gameResult := entity.GameResult{
		UserID:            userID,
		GameStatus:        gameStatus,
		TransactionSource: transactionSource,
		TransactionID:     transactionID,
		Amount:            amount,
		CreatedAt:         time.Now(),
	}

	// Perform the whole operation inside a db transaction
	err := dm.querier.WithTransaction(ctx, func(txn *sqlx.Tx) error {
		if err := dm.persistGameResultTransaction(ctx, txn, userID, &gameResult, balance); err != nil {
			log.Printf("error persisting game result: %v", err)
			return err
		}

		// Commit the transaction
		// Success, continue with the transaction commit
		return nil
	})
	if err != nil {
		log.Printf("error performing game result db transaction: %v", err)
		return nil, entity.ErrCreatingGameResult
	}

	return &gameResult, nil
}

// validateTransaction validates the transaction
// It returns the user if the transaction is valid
func (dm *accountDAO) validateTransaction(ctx context.Context, userID int, gameStatus entity.GameStatus, amount float64, transactionID string) (*entity.User, error) {
	exists, err := dm.querier.TransactionIDExist(ctx, transactionID)
	if err != nil {
		log.Printf("error locating transaction: %v", err)
		return nil, err
	}
	if exists {
		return nil, entity.ErrTransactionIdExists
	}

	user, err := dm.querier.SelectUser(ctx, userID)
	if err != nil {
		log.Printf("error locating user: %v", err)
		return nil, err
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	// No negative balance allowed
	if gameStatus == entity.GameStatusLost && user.Balance < amount {
		return nil, entity.ErrUserNegativeBalance
	}

	return user, nil
}

// calculateNewBalance calculates the new balance based on the game status
func (dm *accountDAO) calculateNewBalance(currentBalance float64, gameStatus entity.GameStatus, amount float64) float64 {
	if gameStatus == entity.GameStatusWin {
		return currentBalance + amount
	}
	return currentBalance - amount
}

// persistGameResultTransaction persists the game result transaction
func (dm *accountDAO) persistGameResultTransaction(ctx context.Context, txn *sqlx.Tx, userID int, gameResult *entity.GameResult, balance float64) error {

	id, err := dm.querier.InsertGameResult(ctx, *txn, *gameResult)
	if err != nil {
		return fmt.Errorf("inserting game result: %w", err)
	}
	gameResult.ID = id

	if err := dm.querier.UpdateUserBalance(ctx, *txn, userID, balance); err != nil {
		return fmt.Errorf("updating user balance: %w", err)
	}

	return nil
}

// RetrieveUser returns the user with the given ID
func (dm *accountDAO) RetrieveUser(ctx context.Context, userID int) (*entity.User, error) {

	user, err := dm.querier.SelectUser(ctx, userID)
	if err != nil {
		log.Printf("error locating user: %v", err)
		return nil, err
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	return user, nil
}
