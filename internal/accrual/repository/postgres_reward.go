package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
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
	query, args, err := psql.
		Insert("accrual.reward_rules").
		Columns("match", "reward", "reward_type").
		Values(rule.Match, rule.Reward, string(rule.RewardType)).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresRewardRepo) GetAll(ctx context.Context) ([]model.RewardRule, error) {
	query, args, err := psql.
		Select("match", "reward", "reward_type").
		From("accrual.reward_rules").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []model.RewardRule
	for rows.Next() {
		var rule model.RewardRule
		err := rows.Scan(&rule.Match, &rule.Reward, &rule.RewardType)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *PostgresRewardRepo) ExistsByMatch(ctx context.Context, match string) (bool, error) {
	query, args, err := psql.
		Select("1").
		From("accrual.reward_rules").
		Where(squirrel.Eq{"match": match}).
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
