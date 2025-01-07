package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/ildomm/account-balance-manager/entity"
	"github.com/ildomm/account-balance-manager/test_helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestHealthHandlerSuccess tests the Health function for a successful response.
func TestHealthHandlerOnSuccess(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	require.NoError(t, err)

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server := Server{}
		server.HealthHandler(w, r)
	})

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err)

	var actual HealthResponse
	err = json.Unmarshal(body, &actual)
	require.NoError(t, err)

	expected := HealthResponse{Status: "ok", Version: "v1"}
	assert.Equal(t, expected, actual)
}

// TestGameResultFuncSuccess tests the CreateGameResultFunc for a successful response using a real server.
func TestGameResultFuncOnSuccess(t *testing.T) {
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

	server.WithAccountManager(daoMock)
	port, err := getAvailablePort(8000)
	assert.NoError(t, err)
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

func TestCreateGameResultFuncOnErrors(t *testing.T) {
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
			server.WithAccountManager(daoMock)
			port, err := getAvailablePort(8000)
			assert.NoError(t, err)
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

// TestRetrieveUserFuncOnSuccess tests the RetrieveUserFunc for a successful response using a real server.
func TestRetrieveUserFuncOnSuccess(t *testing.T) {
	daoMock := test_helpers.NewDAOMock()

	// Set up mock expectations
	testUser := &entity.User{
		ID:      1,
		Balance: 100,
	}

	daoMock.On("RetrieveUser", mock.Anything, mock.Anything).Return(testUser, nil)

	// Create the server and set the mock manager
	server := NewServer()
	server.WithAccountManager(daoMock)
	port, err := getAvailablePort(8000)
	assert.NoError(t, err)
	server.WithListenAddress(port)

	go func() {
		err := server.Run()
		assert.NoError(t, err, "server failed to run")
	}()

	// Create the request
	url := fmt.Sprintf("http://localhost:%d/user/%d/balance", port, testUser.ID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	// Use httptest to create a server
	testServer := httptest.NewServer(server.router())
	defer testServer.Close()

	// Execute the request
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "request to server failed")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var actual UserResponse
	err = json.NewDecoder(resp.Body).Decode(&actual)
	require.NoError(t, err)

	expectedBalance := strconv.FormatFloat(testUser.Balance, 'f', 2, 64)
	assert.Equal(t, expectedBalance, actual.Balance)
	assert.Equal(t, testUser.ID, actual.UserID)
}

func TestRetrieveUserFuncOnErrors(t *testing.T) {
	type testCase struct {
		name           string
		mockSetup      func(daoMock *test_helpers.DAOMock)
		userID         string
		expectedStatus int
	}

	testCases := []testCase{
		{
			name:           "Invalid User ID",
			mockSetup:      nil, // No mock setup required for this test
			userID:         "invalid-user-id",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "User Not Found",
			mockSetup: func(daoMock *test_helpers.DAOMock) {
				daoMock.On("RetrieveUser", mock.Anything, mock.Anything).Return(nil, entity.ErrUserNotFound)
			},
			userID:         "1",
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Internal error",
			mockSetup: func(daoMock *test_helpers.DAOMock) {
				daoMock.On("RetrieveUser", mock.Anything, mock.Anything).Return(nil, errors.New("server error"))
			},
			userID:         "1",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			daoMock := test_helpers.NewDAOMock()
			if tc.mockSetup != nil {
				tc.mockSetup(daoMock)
			}

			server := NewServer()
			server.WithAccountManager(daoMock)
			port, err := getAvailablePort(8000)
			assert.NoError(t, err)
			server.WithListenAddress(port)

			go func() {
				err := server.Run()
				require.NoError(t, err)
			}()

			// Prepare request
			url := fmt.Sprintf("http://localhost:%d/user/%s/balance", port, tc.userID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

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
