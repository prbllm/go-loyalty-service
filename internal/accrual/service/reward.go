package service

import (
	"context"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/accrual/repository"
)

// RewardService отвечает за бизнес-логику, связанную с правилами вознаграждений
type RewardService interface {
	RegisterReward(ctx context.Context, match string, reward int64, rewardType model.RewardType) error
}

// rewardService — реализация RewardService
type rewardService struct {
	rewardRepo repository.RewardRepository
}

// NewRewardService создаёт новый экземпляр RewardService
func NewRewardService(rewardRepo repository.RewardRepository) RewardService {
	return &rewardService{
		rewardRepo: rewardRepo,
	}
}

func (s *rewardService) RegisterReward(ctx context.Context, match string, reward int64, rewardType model.RewardType) error {
	panic("not implemented")
}
