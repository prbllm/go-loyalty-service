package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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
	ErrUserAlreadyExists = errors.New("user with this login already exists")
)

type PostgresRepository struct {
	db     *sql.DB
	logger logger.Logger

	createUserStmt *sql.Stmt
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
	createUserStmt, err := r.db.PrepareContext(ctx,
		"INSERT INTO gophermart.users (login, password_hash) VALUES ($1, $2) RETURNING id")
	if err != nil {
		return fmt.Errorf("failed to prepare createUser statement: %w", err)
	}
	r.createUserStmt = createUserStmt

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
	return nil, errors.New("not implemented")
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	return nil, errors.New("not implemented")
}

func (r *PostgresRepository) CreateOrder(ctx context.Context, userID int64, orderNumber string) error {
	return errors.New("not implemented")
}

func (r *PostgresRepository) GetOrderByNumber(ctx context.Context, orderNumber string) (*model.Order, error) {
	return nil, errors.New("not implemented")
}

func (r *PostgresRepository) GetOrdersByUserID(ctx context.Context, userID int64) ([]*model.Order, error) {
	return nil, errors.New("not implemented")
}

func (r *PostgresRepository) GetOrdersByStatus(ctx context.Context, status string) ([]*model.Order, error) {
	return nil, errors.New("not implemented")
}

func (r *PostgresRepository) UpdateOrderStatus(ctx context.Context, orderNumber string, status string, accrual float64) error {
	return errors.New("not implemented")
}

func (r *PostgresRepository) GetBalance(ctx context.Context, userID int64) (*model.Balance, error) {
	return nil, errors.New("not implemented")
}

func (r *PostgresRepository) WithdrawBalance(ctx context.Context, userID int64, orderNumber string, amount float64) error {
	return errors.New("not implemented")
}

func (r *PostgresRepository) GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	return nil, errors.New("not implemented")
}

func (r *PostgresRepository) AddAccrual(ctx context.Context, userID int64, amount float64) error {
	return errors.New("not implemented")
}
