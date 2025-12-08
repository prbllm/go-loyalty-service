package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

const (
	// PostgreSQL error codes
	ErrCodeUniqueViolation = "23505" // unique_violation

	// Default values
	DefaultUserID = int64(0)
)

var (
	ErrUserAlreadyExists  = errors.New("user with this login already exists")
	ErrOrderAlreadyExists = errors.New("order with this number already exists")
	ErrInsufficientFunds  = errors.New("insufficient funds")
)

type PostgresRepository struct {
	db     *sql.DB
	logger logger.Logger

	createUserStmt     *sql.Stmt
	getUserByLoginStmt *sql.Stmt
	getUserByIDStmt    *sql.Stmt

	createOrderStmt       *sql.Stmt
	getOrderByNumberStmt  *sql.Stmt
	getOrdersByUserIDStmt *sql.Stmt
	getOrdersByStatusStmt *sql.Stmt
}

func NewPostgresRepository(ctx context.Context, dsn string, appLogger logger.Logger) (Repository, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database address cannot be empty")
	}

	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	db := stdlib.OpenDB(*config)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	appLogger.Infof("Connected to PostgreSQL database: %s", dsn)

	if err := runMigrations(db, appLogger); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	repo := &PostgresRepository{
		db:     db,
		logger: appLogger,
	}

	if err := repo.prepareStatements(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	return repo, nil
}

func (r *PostgresRepository) prepareStatements(ctx context.Context) error {
	var prepared []*sql.Stmt
	cleanup := func() {
		for _, stmt := range prepared {
			if stmt != nil {
				if err := stmt.Close(); err != nil {
					r.logger.Errorf("Failed to close prepared statement: %v", err)
				}
			}
		}
	}

	createUserStmt, err := r.db.PrepareContext(ctx,
		"INSERT INTO gophermart.users (login, password_hash) VALUES ($1, $2) RETURNING id")
	if err != nil {
		cleanup()
		return fmt.Errorf("failed to prepare createUser statement: %w", err)
	}
	r.createUserStmt = createUserStmt
	prepared = append(prepared, createUserStmt)

	getUserByLoginStmt, err := r.db.PrepareContext(ctx,
		"SELECT id, login, password_hash, balance, withdrawn, created_at FROM gophermart.users WHERE login = $1")
	if err != nil {
		cleanup()
		return fmt.Errorf("failed to prepare getUserByLogin statement: %w", err)
	}
	r.getUserByLoginStmt = getUserByLoginStmt
	prepared = append(prepared, getUserByLoginStmt)

	getUserByIDStmt, err := r.db.PrepareContext(ctx,
		"SELECT id, login, password_hash, balance, withdrawn, created_at FROM gophermart.users WHERE id = $1")
	if err != nil {
		cleanup()
		return fmt.Errorf("failed to prepare getUserByID statement: %w", err)
	}
	r.getUserByIDStmt = getUserByIDStmt
	prepared = append(prepared, getUserByIDStmt)

	createOrderStmt, err := r.db.PrepareContext(ctx,
		"INSERT INTO gophermart.orders (user_id, number) VALUES ($1, $2) RETURNING id")
	if err != nil {
		cleanup()
		return fmt.Errorf("failed to prepare createOrder statement: %w", err)
	}
	r.createOrderStmt = createOrderStmt
	prepared = append(prepared, createOrderStmt)

	getOrderByNumberStmt, err := r.db.PrepareContext(ctx,
		"SELECT id, user_id, number, status, accrual, uploaded_at FROM gophermart.orders WHERE number = $1")
	if err != nil {
		cleanup()
		return fmt.Errorf("failed to prepare getOrderByNumber statement: %w", err)
	}
	r.getOrderByNumberStmt = getOrderByNumberStmt
	prepared = append(prepared, getOrderByNumberStmt)

	getOrdersByUserIDStmt, err := r.db.PrepareContext(ctx,
		"SELECT id, user_id, number, status, accrual, uploaded_at FROM gophermart.orders WHERE user_id = $1 ORDER BY uploaded_at DESC")
	if err != nil {
		cleanup()
		return fmt.Errorf("failed to prepare getOrdersByUserID statement: %w", err)
	}
	r.getOrdersByUserIDStmt = getOrdersByUserIDStmt
	prepared = append(prepared, getOrdersByUserIDStmt)

	getOrdersByStatusStmt, err := r.db.PrepareContext(ctx,
		"SELECT id, user_id, number, status, accrual, uploaded_at FROM gophermart.orders WHERE status = $1 ORDER BY uploaded_at DESC")
	if err != nil {
		cleanup()
		return fmt.Errorf("failed to prepare getOrdersByStatus statement: %w", err)
	}
	r.getOrdersByStatusStmt = getOrdersByStatusStmt
	prepared = append(prepared, getOrdersByStatusStmt)

	return nil
}

func getMigrationsPath() (string, error) {
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		migrationsDir := filepath.Join(execDir, "migrations", "gophermart")
		if info, err := os.Stat(migrationsDir); err == nil && info.IsDir() {
			absPath, err := filepath.Abs(migrationsDir)
			if err == nil {
				return "file://" + absPath, nil
			}
		}
	}

	wd, err := os.Getwd()
	if err == nil {
		migrationsDir := filepath.Join(wd, "migrations", "gophermart")
		if info, err := os.Stat(migrationsDir); err == nil && info.IsDir() {
			absPath, err := filepath.Abs(migrationsDir)
			if err == nil {
				return "file://" + absPath, nil
			}
		}
	}

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			migrationsDir := filepath.Join(dir, "migrations", "gophermart")
			if info, err := os.Stat(migrationsDir); err == nil && info.IsDir() {
				absPath, err := filepath.Abs(migrationsDir)
				if err == nil {
					return "file://" + absPath, nil
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("failed to find migrations directory: tried executable dir, working dir, and go.mod search")
}

func runMigrations(db *sql.DB, logger logger.Logger) error {
	logger.Info("Running database migrations...")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrationsPath, err := getMigrationsPath()
	if err != nil {
		return fmt.Errorf("failed to get migrations path: %w", err)
	}
	logger.Infof("Migrations path: %s", migrationsPath)

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("No migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	logger.Info("Migrations applied successfully")
	return nil
}

func (r *PostgresRepository) Close() error {
	if r.db == nil {
		return nil
	}

	if r.createUserStmt != nil {
		if err := r.createUserStmt.Close(); err != nil {
			r.logger.Errorf("Failed to close createUser statement: %v", err)
		}
	}

	if r.getUserByLoginStmt != nil {
		if err := r.getUserByLoginStmt.Close(); err != nil {
			r.logger.Errorf("Failed to close getUserByLogin statement: %v", err)
		}
	}

	if r.getUserByIDStmt != nil {
		if err := r.getUserByIDStmt.Close(); err != nil {
			r.logger.Errorf("Failed to close getUserByID statement: %v", err)
		}
	}

	if r.createOrderStmt != nil {
		if err := r.createOrderStmt.Close(); err != nil {
			r.logger.Errorf("Failed to close createOrder statement: %v", err)
		}
	}

	if r.getOrderByNumberStmt != nil {
		if err := r.getOrderByNumberStmt.Close(); err != nil {
			r.logger.Errorf("Failed to close getOrderByNumber statement: %v", err)
		}
	}

	if r.getOrdersByUserIDStmt != nil {
		if err := r.getOrdersByUserIDStmt.Close(); err != nil {
			r.logger.Errorf("Failed to close getOrdersByUserID statement: %v", err)
		}
	}

	if r.getOrdersByStatusStmt != nil {
		if err := r.getOrdersByStatusStmt.Close(); err != nil {
			r.logger.Errorf("Failed to close getOrdersByStatus statement: %v", err)
		}
	}

	return r.db.Close()
}

func (r *PostgresRepository) CreateUser(ctx context.Context, login string, passwordHash string) (int64, error) {
	var id int64
	err := r.createUserStmt.QueryRowContext(ctx, login, passwordHash).Scan(&id)

	if err != nil {
		if isUniqueViolationError(err) {
			r.logger.Errorf("Failed to create user: duplicate login %s", login)
			return DefaultUserID, ErrUserAlreadyExists
		}

		r.logger.Errorf("Failed to create user: %v", err)
		return DefaultUserID, fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Debugf("User created successfully with ID: %d, login: %s", id, login)
	return id, nil
}

func isUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == ErrCodeUniqueViolation
	}
	return false
}

func (r *PostgresRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	var (
		id           int64
		dbLogin      string
		passwordHash string
		balance      float64
		withdrawn    float64
		createdAt    time.Time
	)

	err := r.getUserByLoginStmt.QueryRowContext(ctx, login).Scan(&id, &dbLogin, &passwordHash, &balance, &withdrawn, &createdAt)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:           id,
		Login:        dbLogin,
		PasswordHash: passwordHash,
		Balance:      balance,
		Withdrawn:    withdrawn,
		CreatedAt:    createdAt,
	}, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var (
		dbID         int64
		login        string
		passwordHash string
		balance      float64
		withdrawn    float64
		createdAt    time.Time
	)

	err := r.getUserByIDStmt.QueryRowContext(ctx, id).Scan(&dbID, &login, &passwordHash, &balance, &withdrawn, &createdAt)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:           dbID,
		Login:        login,
		PasswordHash: passwordHash,
		Balance:      balance,
		Withdrawn:    withdrawn,
		CreatedAt:    createdAt,
	}, nil
}

func (r *PostgresRepository) CreateOrder(ctx context.Context, userID int64, orderNumber string) error {
	var id int64
	err := r.createOrderStmt.QueryRowContext(ctx, userID, orderNumber).Scan(&id)
	if err != nil {
		if isUniqueViolationError(err) {
			r.logger.Errorf("Failed to create order: duplicate number %s", orderNumber)
			return ErrOrderAlreadyExists
		}
		return fmt.Errorf("failed to create order: %w", err)
	}
	r.logger.Debugf("Order created successfully with ID: %d, number: %s", id, orderNumber)
	return nil
}

func (r *PostgresRepository) GetOrderByNumber(ctx context.Context, orderNumber string) (*model.Order, error) {
	var (
		id       int64
		userID   int64
		number   string
		status   string
		accrual  float64
		uploaded time.Time
	)

	err := r.getOrderByNumberStmt.QueryRowContext(ctx, orderNumber).Scan(&id, &userID, &number, &status, &accrual, &uploaded)
	if err != nil {
		return nil, err
	}

	return &model.Order{
		ID:         id,
		UserID:     userID,
		Number:     number,
		Status:     status,
		Accrual:    accrual,
		UploadedAt: uploaded,
	}, nil
}

func (r *PostgresRepository) GetOrdersByUserID(ctx context.Context, userID int64) ([]*model.Order, error) {
	rows, err := r.getOrdersByUserIDStmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *PostgresRepository) GetOrdersByStatus(ctx context.Context, status string) ([]*model.Order, error) {
	rows, err := r.getOrdersByStatusStmt.QueryContext(ctx, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *PostgresRepository) UpdateOrderStatus(ctx context.Context, orderNumber string, status string, accrual float64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var (
		userID       int64
		currentState string
	)

	err = tx.QueryRowContext(ctx,
		"SELECT user_id, status FROM gophermart.orders WHERE number = $1 FOR UPDATE",
		orderNumber).Scan(&userID, &currentState)
	if err != nil {
		return err
	}

	if currentState != model.OrderStatusProcessed && status == model.OrderStatusProcessed {
		if _, err := tx.ExecContext(ctx,
			"UPDATE gophermart.users SET balance = balance + $1 WHERE id = $2",
			accrual, userID); err != nil {
			return fmt.Errorf("failed to add accrual to balance: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		"UPDATE gophermart.orders SET status = $1, accrual = $2 WHERE number = $3",
		status, accrual, orderNumber); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetBalance(ctx context.Context, userID int64) (*model.Balance, error) {
	var balance model.Balance

	err := r.db.QueryRowContext(ctx,
		"SELECT balance, withdrawn FROM gophermart.users WHERE id = $1",
		userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

func (r *PostgresRepository) WithdrawBalance(ctx context.Context, userID int64, orderNumber string, amount float64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var (
		balance   float64
		withdrawn float64
	)

	err = tx.QueryRowContext(ctx,
		"SELECT balance, withdrawn FROM gophermart.users WHERE id = $1 FOR UPDATE",
		userID).Scan(&balance, &withdrawn)
	if err != nil {
		return err
	}

	if balance < amount {
		return ErrInsufficientFunds
	}

	if _, err := tx.ExecContext(ctx,
		"UPDATE gophermart.users SET balance = $1, withdrawn = $2 WHERE id = $3",
		balance-amount, withdrawn+amount, userID); err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO gophermart.balance_transactions (user_id, order_number, sum) VALUES ($1, $2, $3)",
		userID, orderNumber, amount); err != nil {
		return fmt.Errorf("failed to insert balance transaction: %w", err)
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT order_number, sum, processed_at FROM gophermart.balance_transactions WHERE user_id = $1 ORDER BY processed_at DESC",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*model.Withdrawal
	for rows.Next() {
		var w model.Withdrawal
		if err := rows.Scan(&w.OrderNumber, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (r *PostgresRepository) AddAccrual(ctx context.Context, userID int64, amount float64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE gophermart.users SET balance = balance + $1 WHERE id = $2",
		amount, userID)
	if err != nil {
		return fmt.Errorf("failed to add accrual: %w", err)
	}

	return nil
}
