package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ildomm/account-balance-manager/entity"
	"github.com/ildomm/account-balance-manager/test_helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestHealthHandlerSuccess tests the Health function for a successful response.
func TestHealthHandlerSuccess(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	require.NoError(t, err)

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server := Server{} // Assuming Server struct exists
		server.HealthHandler(w, r)
	})

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	expected := Response{}
	expected.Data = HealthResponse{Status: "pass", Version: "v1"}
	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err)

	var actual Response
	actual.Data = HealthResponse{}

	err = json.Unmarshal(body, &actual)
	require.NoError(t, err)
}

// TestGameResultFuncSuccess tests the CreateGameResultFunc for a successful response using a real server.
func TestGameResultFuncSuccess(t *testing.T) {
	daoMock := test_helpers.NewDAOMock()

	// Set up mock expectations
	testGameResult := &entity.GameResult{
		ID:            1,
		UserID:        1,
		GameStatus:    "win",
		Amount:        100,
		TransactionID: "123",
		CreatedAt:     time.Now(),
	}
	daoMock.On("CreateGameResult",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(testGameResult, nil)

	// Create the server and set the mock manager
	server := NewServer()
	port := rand.Intn(1000) + 8000
	server.WithAccountManager(daoMock)
	server.WithListenAddress(port)

	go func() {
		err := server.Run()
		assert.NoError(t, err, "server failed to run")
	}()

	// Create the request body
	reqBody := CreateGameResultRequest{
		GameStatus:    "win",
		Amount:        "100",
		TransactionID: "123",
	}
	body, _ := json.Marshal(reqBody)

	// Create the request
	url := fmt.Sprintf("http://localhost:%d/user/%d/transaction", port, testGameResult.UserID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("source-type", string(entity.TransactionSourceGame))

	// Use httptest to create a server
	testServer := httptest.NewServer(server.router())
	defer testServer.Close()

	// Execute the request
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "request to server failed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreateGameResultFunc(t *testing.T) {
	type testCase struct {
		name           string
		mockSetup      func(daoMock *test_helpers.DAOMock)
		userID         string
		requestBody    interface{}
		expectedStatus int
		sourceType     string
	}

	testCases := []testCase{
		{
			name:           "Invalid Request Body",
			mockSetup:      nil, // No mock setup required for this test
			userID:         "1",
			requestBody:    "not a json", // Invalid body
			expectedStatus: http.StatusBadRequest,
			sourceType:     string(entity.TransactionSourceGame),
		},
		{
			name:           "Invalid User ID",
			mockSetup:      nil, // No mock setup required for this test
			userID:         "invalid-user-id",
			requestBody:    CreateGameResultRequest{GameStatus: "win", Amount: "100", TransactionID: "123"},
			expectedStatus: http.StatusBadRequest,
			sourceType:     string(entity.TransactionSourceGame),
		},
		{
			name: "User Not Found",
			mockSetup: func(daoMock *test_helpers.DAOMock) {
				daoMock.On("CreateGameResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, entity.ErrUserNotFound)
			},
			userID:         "1",
			requestBody:    CreateGameResultRequest{GameStatus: "win", Amount: "100", TransactionID: "123"},
			expectedStatus: http.StatusNotFound,
			sourceType:     string(entity.TransactionSourceGame),
		},
		{
			name: "Invalid Game Status",
			mockSetup: func(daoMock *test_helpers.DAOMock) {
				daoMock.On("CreateGameResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, entity.ErrInvalidGameStatus)
			},
			userID:         "1",
			requestBody:    CreateGameResultRequest{GameStatus: "invalid-status", Amount: "100", TransactionID: "123"},
			expectedStatus: http.StatusBadRequest,
			sourceType:     string(entity.TransactionSourceGame),
		},
		{
			name:           "Invalid Transaction Source",
			mockSetup:      nil, // No mock setup required for this test
			userID:         "1",
			requestBody:    CreateGameResultRequest{GameStatus: "win", Amount: "100", TransactionID: "123"},
			expectedStatus: http.StatusBadRequest,
			sourceType:     "invalid-source",
		},
		{
			name:           "Invalid Amount Format",
			mockSetup:      nil, // No mock setup required for this test
			userID:         "1",
			requestBody:    CreateGameResultRequest{GameStatus: "win", Amount: "1b.c23", TransactionID: "123"},
			expectedStatus: http.StatusBadRequest,
			sourceType:     string(entity.TransactionSourceGame),
		},
		{
			name: "Transaction ID Exists",
			mockSetup: func(daoMock *test_helpers.DAOMock) {
				daoMock.On("CreateGameResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, entity.ErrTransactionIdExists)
			},
			userID:         "1",
			requestBody:    CreateGameResultRequest{GameStatus: "win", Amount: "100", TransactionID: "123"},
			expectedStatus: http.StatusNotAcceptable,
			sourceType:     string(entity.TransactionSourceGame),
		},
		{
			name: "User Negative Balance",
			mockSetup: func(daoMock *test_helpers.DAOMock) {
				daoMock.On("CreateGameResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, entity.ErrUserNegativeBalance)
			},
			userID:         "1",
			requestBody:    CreateGameResultRequest{GameStatus: "win", Amount: "100", TransactionID: "123"},
			expectedStatus: http.StatusNotAcceptable,
			sourceType:     string(entity.TransactionSourceGame),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			daoMock := test_helpers.NewDAOMock()
			if tc.mockSetup != nil {
				tc.mockSetup(daoMock)
			}

			server := NewServer()
			port := rand.Intn(1000) + 8000
			server.WithAccountManager(daoMock)
			server.WithListenAddress(port)

			go func() {
				err := server.Run()
				require.NoError(t, err)
			}()

			// Prepare request
			body, err := json.Marshal(tc.requestBody)
			if err != nil {
				body = []byte(tc.requestBody.(string)) // handle non-JSON cases
			}
			url := fmt.Sprintf("http://localhost:%d/user/%s/transaction", port, tc.userID)
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("source-type", tc.sourceType)

			// Use httptest for server mocking
			testServer := httptest.NewServer(server.router())
			defer testServer.Close()

			// Execute request and validate response
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}
