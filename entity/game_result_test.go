package entity

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGameStatusScan(t *testing.T) {
	var status GameStatus
	err := status.Scan("win")
	require.NoError(t, err)
	if status != GameStatusWin {
		t.Errorf("Expected %v, got %v", GameStatusWin, status)
	}
}

func TestGameStatusValue(t *testing.T) {
	status := GameStatusWin
	val, err := status.Value()
	require.NoError(t, err)
	if val != "win" {
		t.Errorf("Expected 'win', got %v", val)
	}
}

func TestTransactionSourceScan(t *testing.T) {
	var source TransactionSource
	err := source.Scan("game")
	require.NoError(t, err)
	if source != TransactionSourceGame {
		t.Errorf("Expected %v, got %v", TransactionSourceGame, source)
	}
}

func TestTransactionSourceValue(t *testing.T) {
	source := TransactionSourceGame
	val, err := source.Value()
	require.NoError(t, err)
	if val != "game" {
		t.Errorf("Expected 'game', got %v", val)
	}
}

func TestGameResult(t *testing.T) {
	id := 1
	userID := 1
	gameStatus := GameStatusWin
	transactionSource := TransactionSourceGame
	transactionID := "tx123"
	amount := 10.15
	createdAt := time.Now()

	gameResult := GameResult{
		ID:                id,
		UserID:            userID,
		GameStatus:        gameStatus,
		TransactionSource: transactionSource,
		TransactionID:     transactionID,
		Amount:            amount,
		CreatedAt:         createdAt,
	}

	if gameResult.ID != id {
		t.Errorf("Expected ID %v, got %v", id, gameResult.ID)
	}
	if gameResult.UserID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, gameResult.UserID)
	}
	if gameResult.GameStatus != gameStatus {
		t.Errorf("Expected GameStatus %v, got %v", gameStatus, gameResult.GameStatus)
	}
	if gameResult.TransactionSource != transactionSource {
		t.Errorf("Expected TransactionSource %v, got %v", transactionSource, gameResult.TransactionSource)
	}
	if gameResult.TransactionID != transactionID {
		t.Errorf("Expected TransactionID %v, got %v", transactionID, gameResult.TransactionID)
	}
	if gameResult.Amount != amount {
		t.Errorf("Expected Amount %v, got %v", amount, gameResult.Amount)
	}
	if !gameResult.CreatedAt.Equal(createdAt) {
		t.Errorf("Expected CreatedAt %v, got %v", createdAt, gameResult.CreatedAt)
	}
}

func TestParseTransactionSource(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  TransactionSource
	}{
		{"Game Source", "game", TransactionSourceGame},
		{"Server Source", "server", TransactionSourceServer},
		{"Payment Source", "payment", TransactionSourcePayment},
		{"Invalid Source", "invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTransactionSource(tt.input.(string))

			if got != nil && tt.want != *got {
				t.Errorf("ParseTransactionSource() = %v, want %v", *got, tt.want)
			}
		})
	}
}
