package server

import "github.com/ildomm/account-balance-manager/entity"

type CreateGameResultRequest struct {
	GameStatus    entity.GameStatus `json:"state"`
	Amount        string            `json:"amount"` // TODO: This could be float64
	TransactionID string            `json:"transactionId"`
}
