package entity

import (
	"errors"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserNegativeBalance = errors.New("negative balance not allowed")
var ErrTransactionIdExists = errors.New("transaction id already exists")
var ErrInvalidGameStatus = errors.New("invalid game status")
var ErrRequestPayload = errors.New("invalid request body")
var ErrInvalidAmount = errors.New("invalid amount format")
var ErrInvalidUser = errors.New("invalid user Id")
var ErrInvalidTransactionSource = errors.New("invalid transaction source")
var ErrCreatingGameResult = errors.New("error recording game result")
var ErrServerInternal = errors.New("internal server error")
