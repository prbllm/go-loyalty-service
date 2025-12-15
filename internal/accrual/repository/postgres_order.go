package repository

import (
	"context"
	"database/sql"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
)

// PostgresOrderRepo реализует OrderRepository с использованием PostgreSQL
type PostgresOrderRepo struct {
	db *sql.DB
}

func NewPostgresOrderRepo(db *sql.DB) *PostgresOrderRepo {
	return &PostgresOrderRepo{db: db}
}

func (r *PostgresOrderRepo) Create(ctx context.Context, order *model.Order) error {
	panic("not implemented")
}

func (r *PostgresOrderRepo) GetByNumber(ctx context.Context, number string) (*model.Order, error) {
	panic("not implemented")
}

func (r *PostgresOrderRepo) UpdateStatusAndAccrual(ctx context.Context, number string, status model.OrderStatus, accrual *int64) error {
	panic("not implemented")
}

func (r *PostgresOrderRepo) GetAllProcessing(ctx context.Context) ([]model.Order, error) {
	panic("not implemented")
}
