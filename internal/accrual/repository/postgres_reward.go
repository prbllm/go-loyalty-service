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

func (r *PostgresRewardRepo) Create(ctx context.Context, rule model.RewardRule) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO reward_rules (match, reward, reward_type) VALUES ($1, $2, $3)", rule.Match, rule.Reward, rule.RewardType)
	return err
}

func (r *PostgresRewardRepo) GetAll(ctx context.Context) ([]model.RewardRule, error) {
	panic("not implemented")
}

func (r *PostgresRewardRepo) ExistsByMatch(ctx context.Context, match string) (bool, error) {
	row := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM reward_rules WHERE match = $1)", match)

	var exists bool
	err := row.Scan(&exists)

	return exists, err
}
