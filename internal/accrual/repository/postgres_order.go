package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

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

	query, args, err := psql.
		Insert("accrual.orders").
		Columns("number", "status", "accrual", "goods").
		Values(order.Number, string(order.Status), order.Accrual, goodsData).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresOrderRepo) IsOrderExists(ctx context.Context, number string) (bool, error) {
	query, args, err := psql.
		Select("1").
		From("accrual.orders").
		Where(squirrel.Eq{"number": number}).
		Limit(1).
		ToSql()
	if err != nil {
		return false, err
	}
	var dummy int
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&dummy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *PostgresOrderRepo) GetByNumber(ctx context.Context, number string) (*model.Order, error) {
	query, args, err := psql.
		Select("status", "accrual", "goods").
		From("accrual.orders").
		Where(squirrel.Eq{"number": number}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, query, args...)

	var goodsData []byte
	order := model.Order{Number: number}
	err = row.Scan(&order.Status, &order.Accrual, &goodsData)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(goodsData, &order.Goods)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *PostgresOrderRepo) UpdateStatusAndAccrual(ctx context.Context, number string, status model.OrderStatus, accrual *int64) error {
	query, args, err := psql.
		Update("accrual.orders").
		Set("status", string(status)).
		Set("accrual", accrual).
		Where(squirrel.Eq{"number": number}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}
