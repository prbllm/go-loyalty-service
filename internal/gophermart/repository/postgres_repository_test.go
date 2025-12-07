package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const (
	defaultTestAdminDSN = "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	defaultTestDBName   = "loyalty_test"
	defaultTestDSN      = "postgresql://postgres:postgres@localhost:5432/loyalty_test?sslmode=disable"
)

func setupTestDB(t *testing.T) (*PostgresRepository, func()) {
	ctx := context.Background()
	appLogger := zaptest.NewLogger(t).Sugar()

	adminDSN := defaultTestAdminDSN

	adminConfig, err := pgx.ParseConfig(adminDSN)
	require.NoError(t, err, "Failed to parse admin DSN")

	adminDB := stdlib.OpenDB(*adminConfig)
	defer adminDB.Close()

	err = adminDB.PingContext(ctx)
	require.NoError(t, err, "Failed to connect to admin database")

	var dbExists int
	err = adminDB.QueryRowContext(ctx,
		"SELECT 1 FROM pg_database WHERE datname = $1", defaultTestDBName).Scan(&dbExists)

	if err == sql.ErrNoRows {
		_, err = adminDB.ExecContext(ctx, "CREATE DATABASE "+defaultTestDBName)
		require.NoError(t, err, "Failed to create test database")
		appLogger.Infof("Created test database: %s", defaultTestDBName)
	} else {
		require.NoError(t, err, "Failed to check if test database exists")
	}

	dsn := os.Getenv("TEST_DATABASE_URI")
	if dsn == "" {
		dsn = defaultTestDSN
	}

	repo, err := NewPostgresRepository(ctx, dsn, appLogger)
	require.NoError(t, err)

	cleanup := func() {
		pgRepo := repo.(*PostgresRepository)
		if pgRepo.db != nil {
			_, err := pgRepo.db.ExecContext(ctx, "TRUNCATE TABLE gophermart.users CASCADE")
			if err != nil {
				t.Logf("Failed to truncate users table: %v", err)
			}
		}
		if err := pgRepo.Close(); err != nil {
			t.Logf("Failed to close database: %v", err)
		}
	}

	return repo.(*PostgresRepository), cleanup
}

func TestCreateUser_Success(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	login := "testuser"
	passwordHash := "$2a$10$testhash"

	id, err := repo.CreateUser(ctx, login, passwordHash)

	assert.NoError(t, err)
	assert.Greater(t, id, DefaultUserID)

	var dbLogin string
	var dbPasswordHash string
	var dbID int64
	err = repo.db.QueryRowContext(ctx,
		"SELECT id, login, password_hash FROM gophermart.users WHERE id = $1",
		id).Scan(&dbID, &dbLogin, &dbPasswordHash)

	assert.NoError(t, err)
	assert.Equal(t, id, dbID)
	assert.Equal(t, login, dbLogin)
	assert.Equal(t, passwordHash, dbPasswordHash)
}

func TestCreateUser_DuplicateLogin(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	login := "duplicateuser"
	passwordHash := "$2a$10$testhash"

	id1, err := repo.CreateUser(ctx, login, passwordHash)
	assert.NoError(t, err)
	assert.Greater(t, id1, DefaultUserID)

	id2, err := repo.CreateUser(ctx, login, passwordHash)
	assert.Error(t, err)
	assert.Equal(t, DefaultUserID, id2)
	assert.Equal(t, ErrUserAlreadyExists, err)
}
