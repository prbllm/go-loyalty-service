package service

import (
	"context"
	"errors"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/accrual/repository"
)

// RewardService отвечает за бизнес-логику, связанную с правилами вознаграждений
type RewardService interface {
	RegisterReward(ctx context.Context, reward model.RewardRule) error
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

var ErrMatchAlreadyExists = errors.New("match already exists")

func (s *rewardService) RegisterReward(ctx context.Context, reward model.RewardRule) error {
	// Проверяем, существует ли правило с таким match
	exists, err := s.rewardRepo.ExistsByMatch(ctx, reward.Match)
	if err != nil {
		return err
	}

	if exists {
		return ErrMatchAlreadyExists
	}

	// Сохраняем правило
	err = s.rewardRepo.Create(ctx, reward)
	if err != nil {
		return err
	}

	return nil
}
