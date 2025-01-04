package database

import (
	"context"
	"fmt"
	"github.com/ildomm/account-balance-manager/entity"
	"github.com/ildomm/account-balance-manager/test_helpers"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPostgresQuerier(t *testing.T) {
	testDB := test_helpers.NewTestDatabase(t)
	dbURL := testDB.ConnectionString(t) + "?sslmode=disable"

	ctx := context.Background()

	t.Run("NewPostgresQuerier_Success", func(t *testing.T) {
		querier, err := NewPostgresQuerier(ctx, dbURL)
		require.NoError(t, err)
		require.NotNil(t, querier)

		defer querier.Close()

		assert.NotNil(t, querier.dbConn)

		// Check the number of migration files in the folder
		migrationFiles, err := fs.ReadDir("migrations")
		require.NoError(t, err)

		// Count the number of migration files
		expectedNumMigrations := len(migrationFiles) / 2 // Each migration file has a corresponding .down.sql file

		// Query the "schema_migrations" table to get the version
		var version string
		err = querier.dbConn.Get(&version, "SELECT version FROM schema_migrations")
		require.NoError(t, err)

		// Convert the version to an integer
		versionInt, err := strconv.Atoi(strings.TrimSpace(version))
		require.NoError(t, err)

		// Compare the number of migrations with the version in the database
		assert.Equal(t, expectedNumMigrations, versionInt, fmt.Sprintf("Number of migrations should match the version in the database. Expected: %d, Actual: %d", expectedNumMigrations, versionInt))
	})

	t.Run("NewPostgresQuerier_InvalidURL", func(t *testing.T) {
		_, err := NewPostgresQuerier(ctx, "invalid-url")
		require.Error(t, err)
	})
}

func setupTestQuerier(t *testing.T) (context.Context, func(t *testing.T), *PostgresQuerier) {
	testDB := test_helpers.NewTestDatabase(t)
	ctx := context.Background()
	q, err := NewPostgresQuerier(ctx, testDB.ConnectionString(t)+"?sslmode=disable")
	require.NoError(t, err)

	return ctx, func(t *testing.T) {
		testDB.Close(t)
	}, q
}

func TestDatabaseWithTransaction(t *testing.T) {
	ctx, teardownTest, q := setupTestQuerier(t)
	defer teardownTest(t)

	t.Run("InsertGameResult_Success", func(t *testing.T) {
		gameResult := entity.GameResult{
			UserID:            1,
			GameStatus:        entity.GameStatusWin,
			TransactionSource: entity.TransactionSourceServer,
			TransactionID:     "anything",
			Amount:            10,
			CreatedAt:         time.Now(),
		}

		// Start a transaction that is expected to WORK
		err := q.WithTransaction(ctx, func(txn *sqlx.Tx) error {

			id, err := q.InsertGameResult(ctx, *txn, gameResult)
			require.NoError(t, err)

			gameResult.ID = id

			// No error, then the db commit() will happen
			return nil
		})
		require.NoError(t, err)
	})
}

func TestDatabaseBasicOperations(t *testing.T) {
	ctx, teardownTest, q := setupTestQuerier(t)
	defer teardownTest(t)

	userID := 1

	t.Run("SelectUser_Success", func(t *testing.T) {
		user, err := q.SelectUser(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, userID, user.ID)
	})

	t.Run("TransactionIDExist_None", func(t *testing.T) {
		exist, err := q.TransactionIDExist(ctx, "anything")
		require.NoError(t, err)
		require.False(t, exist)
	})

	t.Run("TransactionIDExist_One", func(t *testing.T) {
		gameResult := entity.GameResult{
			UserID:            1,
			GameStatus:        entity.GameStatusWin,
			TransactionSource: entity.TransactionSourceServer,
			TransactionID:     "anything",
			Amount:            10,
			CreatedAt:         time.Now(),
		}

		// Start a transaction that is expected to WORK
		err := q.WithTransaction(ctx, func(txn *sqlx.Tx) error {

			id, err := q.InsertGameResult(ctx, *txn, gameResult)
			require.NoError(t, err)

			gameResult.ID = id

			// No error, then the db commit() will happen
			return nil
		})
		require.NoError(t, err)

		exist, err := q.TransactionIDExist(ctx, "anything")
		require.NoError(t, err)
		require.True(t, exist)
	})
}
