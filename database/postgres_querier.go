package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/ildomm/account-balance-manager/entity"
	"github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/url"
)

type PostgresQuerier struct {
	dbURL  string
	dbConn *sqlx.DB
	ctx    context.Context
}

func NewPostgresQuerier(ctx context.Context, url string) (*PostgresQuerier, error) {
	querier := PostgresQuerier{dbURL: url, ctx: ctx}

	_, err := pgx.ParseConfig(url)
	if err != nil {
		return &querier, err
	}

	// Open the connection using the DataDog wrapper methods
	querier.dbConn, err = sqlx.Open("pgx", url)
	if err != nil {
		return &querier, err
	}
	log.Print("opened database connection")

	// Ping the database to check that the connection is actually working
	err = querier.dbConn.Ping()
	if err != nil {
		return &querier, err
	}

	// Migrate the database
	err = querier.migrate()
	if err != nil {
		return &querier, err
	}
	log.Print("database migration complete")

	return &querier, nil
}

func (q *PostgresQuerier) Close() {
	q.dbConn.Close()
	log.Print("closed database connection")
}

var (
	//go:embed migrations/*.sql
	fs embed.FS
)

func (q *PostgresQuerier) migrate() error {

	// Amend the database URl with custom parameter so that we can specify the
	// table name to be used to hold database migration state
	migrationsURL, err := q.migrationsURL()
	if err != nil {
		return err
	}

	// Load the migrations from our embedded resources
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	// Use a custom table name for schema migrations
	m, err := migrate.NewWithSourceInstance("iofs", d, migrationsURL)
	if err != nil {
		return err
	}

	// Migrate all the way up ...
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

const (
	CustomMigrationParam = "x-migrations-table"
	CustomMigrationValue = "schema_migrations"
)

func (q *PostgresQuerier) migrationsURL() (string, error) {
	url, err := url.Parse(q.dbURL)
	if err != nil {
		return "", err
	}

	// Add the new Query parameter that specifies the table name for the migrations
	values := url.Query()
	values.Add(CustomMigrationParam, CustomMigrationValue)

	// Replace the Query parameters in the original URL & return
	url.RawQuery = values.Encode()
	return url.String(), nil
}

////////////////////////////////// Database Querier standard operations /////////////////////////////////////////////////////////

// WithTransaction creates a new transaction and handles rollback/commit based on the
// error object returned by the `TxFn`
func (q *PostgresQuerier) WithTransaction(ctx context.Context, fn func(*sqlx.Tx) error) (err error) {

	// Starting database transaction
	tx, err := q.dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and re-panic
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			// something went wrong, rollback
			err = tx.Rollback()
		} else {
			// all good, commit
			err = tx.Commit()
		}
	}()

	// The database transaction starts taking effect here
	err = fn(tx)
	return err
}

////////////////////////////////// Database Querier domain operations /////////////////////////////////////////////////////////

const insertGameResultSQL = `
	INSERT INTO game_results ( user_id, game_status, transaction_source, transaction_id, amount, created_at)
	VALUES                   ( $1,      $2,          $3,                 $4,             $5,     $6)
	RETURNING id`

func (q *PostgresQuerier) InsertGameResult(ctx context.Context, txn sqlx.Tx, gameResult entity.GameResult) (int, error) {
	var id int

	err := txn.GetContext(
		ctx,
		&id,
		insertGameResultSQL,
		gameResult.UserID,
		gameResult.GameStatus,
		gameResult.TransactionSource,
		gameResult.TransactionID,
		gameResult.Amount,
		gameResult.CreatedAt)

	return id, err
}

const selectUserSQL = `SELECT * FROM users WHERE id = $1`

func (q *PostgresQuerier) SelectUser(ctx context.Context, userID int) (*entity.User, error) {
	var user entity.User

	err := q.dbConn.GetContext(
		ctx,
		&user,
		selectUserSQL,
		userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &user, nil
}

const selectCheckTransactionSQL = `SELECT count(*) FROM game_results WHERE transaction_id = $1`

func (q *PostgresQuerier) TransactionIDExist(ctx context.Context, transactionID string) (bool, error) {

	row := q.dbConn.QueryRowContext(ctx, selectCheckTransactionSQL, transactionID)
	var count int64

	err := row.Scan(&count)

	if count > 0 {
		return true, err
	} else {
		return false, err
	}
}

const updateUserSQL = `
	UPDATE users
	SET 
		balance = :balance
	WHERE id = :id`

func (q *PostgresQuerier) UpdateUserBalance(ctx context.Context, txn sqlx.Tx, userID int, balance float64) error {
	user := entity.User{
		ID:      userID,
		Balance: balance,
	}

	_, err := txn.NamedExecContext(ctx, updateUserSQL, user)

	return err
}
