package database

import (
	"context"
	"fmt"
	"github.com/ildomm/account-balance-manager/test_helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"strings"
	"testing"
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

func TestDatabaseWithTransaction(t *testing.T) {
	// TODO: Implement
}

func TestDatabaseBasicOperations(t *testing.T) {
	// TODO: Implement
}