package repository

import (
	"context"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
)

//go:generate mockgen -source=reward.go -destination=../../mocks/accrual/reward_repository.go -package=mocks

// RewardRepository отвечает за операции с правилами вознаграждений
type RewardRepository interface {
	// Create создаёт новое правило начисления
	Create(ctx context.Context, rule model.RewardRule) error

	// GetAll возвращает все активные правила
	GetAll(ctx context.Context) ([]model.RewardRule, error)

	// ExistsByMatch проверяет, существует ли правило с указанным match-ключом
	ExistsByMatch(ctx context.Context, match string) (bool, error)
}
