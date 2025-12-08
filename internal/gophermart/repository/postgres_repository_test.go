package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"

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

func TestGetUserByLogin(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	login := "login_lookup"
	passwordHash := "$2a$10$loginlookuphash"

	createdID, err := repo.CreateUser(ctx, login, passwordHash)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		user, err := repo.GetUserByLogin(ctx, login)
		require.NoError(t, err)

		assert.Equal(t, createdID, user.ID)
		assert.Equal(t, login, user.Login)
		assert.Equal(t, passwordHash, user.PasswordHash)
		assert.Equal(t, float64(0), user.Balance)
		assert.Equal(t, float64(0), user.Withdrawn)
		assert.False(t, user.CreatedAt.IsZero())
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetUserByLogin(ctx, "missing_login")
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestGetUserByID(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	login := "id_lookup"
	passwordHash := "$2a$10$idlookuphash"

	createdID, err := repo.CreateUser(ctx, login, passwordHash)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		user, err := repo.GetUserByID(ctx, createdID)
		require.NoError(t, err)

		assert.Equal(t, createdID, user.ID)
		assert.Equal(t, login, user.Login)
		assert.Equal(t, passwordHash, user.PasswordHash)
		assert.Equal(t, float64(0), user.Balance)
		assert.Equal(t, float64(0), user.Withdrawn)
		assert.False(t, user.CreatedAt.IsZero())
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetUserByID(ctx, createdID+999)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestCreateOrder(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userID, err := repo.CreateUser(ctx, "order_user", "$2a$10$orderhash")
	require.NoError(t, err)

	orderNumber := "1234567890"

	err = repo.CreateOrder(ctx, userID, orderNumber)
	require.NoError(t, err)

	var (
		dbID      int64
		dbUserID  int64
		dbNumber  string
		dbStatus  string
		dbAccrual float64
	)
	err = repo.db.QueryRowContext(ctx,
		"SELECT id, user_id, number, status, accrual FROM gophermart.orders WHERE number = $1",
		orderNumber).Scan(&dbID, &dbUserID, &dbNumber, &dbStatus, &dbAccrual)

	require.NoError(t, err)
	assert.Equal(t, userID, dbUserID)
	assert.Equal(t, orderNumber, dbNumber)
	assert.Equal(t, "NEW", dbStatus)
	assert.Equal(t, float64(0), dbAccrual)
	assert.Greater(t, dbID, int64(0))

	err = repo.CreateOrder(ctx, userID, orderNumber)
	require.Error(t, err)
	assert.Equal(t, ErrOrderAlreadyExists, err)
}

func TestGetOrderByNumber(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userID, err := repo.CreateUser(ctx, "order_lookup_user", "$2a$10$orderlookuphash")
	require.NoError(t, err)

	orderNumber := "lookup123"
	err = repo.CreateOrder(ctx, userID, orderNumber)
	require.NoError(t, err)

	order, err := repo.GetOrderByNumber(ctx, orderNumber)
	require.NoError(t, err)

	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, orderNumber, order.Number)
	assert.Equal(t, "NEW", order.Status)
	assert.Equal(t, float64(0), order.Accrual)
	assert.False(t, order.UploadedAt.IsZero())

	_, err = repo.GetOrderByNumber(ctx, "missing_order")
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestGetOrdersByUserID(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userID, err := repo.CreateUser(ctx, "orders_by_user", "$2a$10$ordersbyuserhash")
	require.NoError(t, err)

	otherUserID, err := repo.CreateUser(ctx, "other_user", "$2a$10$otheruserhash")
	require.NoError(t, err)

	numbers := []string{"order1", "order2", "order3"}
	for _, num := range numbers {
		require.NoError(t, repo.CreateOrder(ctx, userID, num))
	}

	require.NoError(t, repo.CreateOrder(ctx, otherUserID, "other_order"))

	orders, err := repo.GetOrdersByUserID(ctx, userID)
	require.NoError(t, err)

	require.Len(t, orders, 3)
	receivedNumbers := []string{orders[0].Number, orders[1].Number, orders[2].Number}
	assert.ElementsMatch(t, numbers, receivedNumbers)

	orders, err = repo.GetOrdersByUserID(ctx, 99999)
	require.NoError(t, err)
	assert.Len(t, orders, 0)
}

func TestGetOrdersByStatus(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userID, err := repo.CreateUser(ctx, "orders_by_status", "$2a$10$ordersbystatushash")
	require.NoError(t, err)

	require.NoError(t, repo.CreateOrder(ctx, userID, "st_order1"))
	require.NoError(t, repo.CreateOrder(ctx, userID, "st_order2"))
	require.NoError(t, repo.CreateOrder(ctx, userID, "st_order3"))

	_, err = repo.db.ExecContext(ctx,
		"UPDATE gophermart.orders SET status = 'PROCESSED', accrual = 10.5 WHERE number = $1",
		"st_order2")
	require.NoError(t, err)

	ordersNew, err := repo.GetOrdersByStatus(ctx, "NEW")
	require.NoError(t, err)
	assert.True(t, len(ordersNew) >= 2)

	ordersProcessed, err := repo.GetOrdersByStatus(ctx, "PROCESSED")
	require.NoError(t, err)
	require.Len(t, ordersProcessed, 1)
	assert.Equal(t, "st_order2", ordersProcessed[0].Number)
	assert.Equal(t, float64(10.5), ordersProcessed[0].Accrual)
}

func TestUpdateOrderStatus_AddsAccrualOnce(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userID, err := repo.CreateUser(ctx, "status_user", "$2a$10$statushash")
	require.NoError(t, err)

	orderNumber := "status_order"
	require.NoError(t, repo.CreateOrder(ctx, userID, orderNumber))

	err = repo.UpdateOrderStatus(ctx, orderNumber, model.OrderStatusProcessed, 15.5)
	require.NoError(t, err)

	var status string
	var accrual float64
	err = repo.db.QueryRowContext(ctx,
		"SELECT status, accrual FROM gophermart.orders WHERE number = $1",
		orderNumber).Scan(&status, &accrual)
	require.NoError(t, err)
	assert.Equal(t, model.OrderStatusProcessed, status)
	assert.Equal(t, float64(15.5), accrual)

	balance, err := repo.GetBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, float64(15.5), balance.Current)

	err = repo.UpdateOrderStatus(ctx, orderNumber, model.OrderStatusProcessed, 20)
	require.NoError(t, err)

	balance, err = repo.GetBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, float64(15.5), balance.Current)

	err = repo.db.QueryRowContext(ctx,
		"SELECT accrual FROM gophermart.orders WHERE number = $1",
		orderNumber).Scan(&accrual)
	require.NoError(t, err)
	assert.Equal(t, float64(20), accrual)
}

func TestGetBalanceAndAddAccrual(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userID, err := repo.CreateUser(ctx, "balance_user", "$2a$10$balancehash")
	require.NoError(t, err)

	balance, err := repo.GetBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, float64(0), balance.Current)
	assert.Equal(t, float64(0), balance.Withdrawn)

	err = repo.AddAccrual(ctx, userID, 12.3)
	require.NoError(t, err)

	balance, err = repo.GetBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, float64(12.3), balance.Current)
	assert.Equal(t, float64(0), balance.Withdrawn)
}

func TestWithdrawBalance(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userID, err := repo.CreateUser(ctx, "withdraw_user", "$2a$10$withdrawhash")
	require.NoError(t, err)

	require.NoError(t, repo.AddAccrual(ctx, userID, 25))

	err = repo.WithdrawBalance(ctx, userID, "wd_order_1", 10)
	require.NoError(t, err)

	balance, err := repo.GetBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, float64(15), balance.Current)
	assert.Equal(t, float64(10), balance.Withdrawn)

	var (
		dbOrder string
		dbSum   float64
	)
	err = repo.db.QueryRowContext(ctx,
		"SELECT order_number, sum FROM gophermart.balance_transactions WHERE user_id = $1",
		userID).Scan(&dbOrder, &dbSum)
	require.NoError(t, err)
	assert.Equal(t, "wd_order_1", dbOrder)
	assert.Equal(t, float64(10), dbSum)
}

func TestWithdrawBalance_InsufficientFunds(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userID, err := repo.CreateUser(ctx, "withdraw_fail_user", "$2a$10$withdrawfailhash")
	require.NoError(t, err)

	err = repo.WithdrawBalance(ctx, userID, "wd_order_fail", 5)
	require.Error(t, err)
	assert.Equal(t, ErrInsufficientFunds, err)
}

func TestGetWithdrawals(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	userID, err := repo.CreateUser(ctx, "withdrawals_user", "$2a$10$withdrawalshash")
	require.NoError(t, err)

	require.NoError(t, repo.AddAccrual(ctx, userID, 30))
	require.NoError(t, repo.WithdrawBalance(ctx, userID, "wd_order_a", 10))
	require.NoError(t, repo.WithdrawBalance(ctx, userID, "wd_order_b", 5))

	withdrawals, err := repo.GetWithdrawals(ctx, userID)
	require.NoError(t, err)
	require.Len(t, withdrawals, 2)
	assert.Equal(t, "wd_order_b", withdrawals[0].OrderNumber)
	assert.Equal(t, float64(5), withdrawals[0].Sum)
	assert.Equal(t, "wd_order_a", withdrawals[1].OrderNumber)
	assert.Equal(t, float64(10), withdrawals[1].Sum)
}
