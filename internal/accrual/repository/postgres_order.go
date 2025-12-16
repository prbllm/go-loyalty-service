package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
)

// PostgresOrderRepo реализует OrderRepository с использованием PostgreSQL
type PostgresOrderRepo struct {
	db *sql.DB
}

func NewPostgresOrderRepo(db *sql.DB) *PostgresOrderRepo {
	return &PostgresOrderRepo{db: db}
}

func (r *PostgresOrderRepo) Create(ctx context.Context, order model.Order) error {
	goodsData, err := json.Marshal(order.Goods)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, "INSERT INTO orders (number, status, accrual, goods) VALUES ($1, $2, $3, $4)", order.Number, order.Status, order.Accrual, goodsData)
	return err
}

func (r *PostgresOrderRepo) IsOrderExists(ctx context.Context, number string) (bool, error) {
	row := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM orders WHERE number = $1)", number)

	var exists bool
	err := row.Scan(&exists)

	return exists, err
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
