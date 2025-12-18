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
	rows, err := r.db.QueryContext(ctx, "SELECT match, reward, reward_type FROM reward_rules")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var rules []model.RewardRule
	for rows.Next() {
		var rule model.RewardRule
		err := rows.Scan(
			&rule.Match,
			&rule.Reward,     // float64
			&rule.RewardType, // RewardType (string)
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	// Проверяем ошибки итерации
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *PostgresRewardRepo) ExistsByMatch(ctx context.Context, match string) (bool, error) {
	row := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM reward_rules WHERE match = $1)", match)

	var exists bool
	err := row.Scan(&exists)

	return exists, err
}
