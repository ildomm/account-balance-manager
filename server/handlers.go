package server

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/ildomm/account-balance-manager/dao"
	"github.com/ildomm/account-balance-manager/entity"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// accountHandler handles all requests related to game results.
type accountHandler struct {
	accountDAO dao.DAO
}

func NewAccountHandler(accountDAO dao.DAO) *accountHandler {
	return &accountHandler{
		accountDAO: accountDAO,
	}
}

// CreateGameResultFunc handles the request to create a new game result.
func (h *accountHandler) CreateGameResultFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate the headers.
	transactionSource := entity.ParseTransactionSource(strings.ToLower(r.Header.Get("Source-Type")))
	if transactionSource == nil {
		WriteErrorResponse(w, http.StatusBadRequest, []string{entity.ErrInvalidTransactionSource.Error()})
		return
	}

	// Validate the request body.
	var req CreateGameResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, []string{entity.ErrRequestPayload.Error()})
		return
	}

	// Basic request validation
	if req.TransactionID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"transaction_id is required"})
		return
	}

	// Validate amount type cast and value
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil || amount <= 0 {
		WriteErrorResponse(w, http.StatusBadRequest, []string{entity.ErrInvalidAmount.Error()})
		return
	}

	// Validate the game status.
	if req.GameStatus != entity.GameStatusWin && req.GameStatus != entity.GameStatusLose {
		WriteErrorResponse(w, http.StatusBadRequest, []string{entity.ErrInvalidGameStatus.Error()})
		return
	}

	// Extract and validate the user ID from the request path.
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil || userID <= 0 {
		WriteErrorResponse(w, http.StatusBadRequest, []string{entity.ErrInvalidUser.Error()})
		return
	}

	// Perform the business logic.
	_, err = h.accountDAO.CreateGameResult(r.Context(), userID, req.GameStatus, amount, *transactionSource, req.TransactionID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrUserNotFound):
			WriteErrorResponse(w, http.StatusNotFound, []string{err.Error()})
		case errors.Is(err, entity.ErrTransactionIdExists) || errors.Is(err, entity.ErrUserNegativeBalance):
			WriteErrorResponse(w, http.StatusNotAcceptable, []string{err.Error()})
		default:
			// Log the actual error but return a generic message
			log.Printf("Internal error: %v", err)
			WriteErrorResponse(w, http.StatusInternalServerError, []string{"An internal error occurred"})
		}
		return
	}

	WriteAPIResponse(w, http.StatusOK, nil)
}

// RetrieveUserFunc handles the request to retrieve the account user.
func (h *accountHandler) RetrieveUserFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract and validate the user ID from the request path.
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil || userID <= 0 {
		WriteErrorResponse(w, http.StatusBadRequest, []string{entity.ErrInvalidUser.Error()})
		return
	}

	user, err := h.accountDAO.RetrieveUser(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrUserNotFound):
			WriteErrorResponse(w, http.StatusNotFound, []string{err.Error()})
		default:
			// Log the actual error but return a generic message
			log.Printf("Internal error: %v", err)
			WriteErrorResponse(w, http.StatusInternalServerError, []string{"An internal error occurred"})
		}
		return
	}

	userResponse := transformUserResponse(*user)
	WriteAPIResponse(w, http.StatusOK, userResponse)
}

// Transform entity.User to server.UserResponse
func transformUserResponse(user entity.User) UserResponse {
	return UserResponse{
		UserID: user.ID,

		// Transform the balance to a string, rounded to 2 decimal places
		Balance: strconv.FormatFloat(user.Balance, 'f', 2, 64),
	}
}
