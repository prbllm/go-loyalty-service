package repository

import (
	"context"
	"database/sql"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
)

// PostgresRewardRepo реализует RewardRepository с использованием PostgreSQL
type PostgresRewardRepo struct {
	db *sql.DB
}

func NewPostgresRewardRepo(db *sql.DB) *PostgresRewardRepo {
	return &PostgresRewardRepo{db: db}
}

func (r *PostgresRewardRepo) Create(ctx context.Context, rule *model.RewardRule) error {
	panic("not implemented")
}

func (r *PostgresRewardRepo) GetAll(ctx context.Context) ([]model.RewardRule, error) {
	panic("not implemented")
}

func (r *PostgresRewardRepo) ExistsByMatch(ctx context.Context, match string) (bool, error) {
	panic("not implemented")
}
