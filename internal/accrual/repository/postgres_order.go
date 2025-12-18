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
	_, err = r.db.ExecContext(ctx, "INSERT INTO accrual.orders (number, status, accrual, goods) VALUES ($1, $2, $3, $4)", order.Number, order.Status, order.Accrual, goodsData)
	return err
}

func (r *PostgresOrderRepo) IsOrderExists(ctx context.Context, number string) (bool, error) {
	row := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM accrual.orders WHERE number = $1)", number)

	var exists bool
	err := row.Scan(&exists)

	return exists, err
}

func (r *PostgresOrderRepo) GetByNumber(ctx context.Context, number string) (model.Order, error) {
	row := r.db.QueryRowContext(ctx, "SELECT status, accrual, goods FROM accrual.orders WHERE number = $1", number)

	var goodsData []byte
	order := model.Order{Number: number}
	err := row.Scan(&order.Status, &order.Accrual, &goodsData)
	if err != nil {
		return model.Order{}, err
	}

	err = json.Unmarshal(goodsData, &order.Goods)
	if err != nil {
		return model.Order{}, err
	}

	return order, err
}

func (r *PostgresOrderRepo) UpdateStatusAndAccrual(ctx context.Context, number string, status model.OrderStatus, accrual *int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accrual.orders SET status = $1, accrual = $2 WHERE number = $3", status, accrual, number)
	return err
}

func (r *PostgresOrderRepo) GetAllProcessing(ctx context.Context) ([]model.Order, error) {
	panic("not implemented")
}
