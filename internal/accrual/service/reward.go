package service

import (
	"context"
	"errors"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/accrual/repository"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

//go:generate mockgen -source=reward.go -destination=../../mocks/accrual/reward_service.go -package=mocks

// RewardService отвечает за бизнес-логику, связанную с правилами вознаграждений
type RewardService interface {
	RegisterReward(ctx context.Context, reward model.RewardRule) error
}

// rewardService — реализация RewardService
type rewardService struct {
	rewardRepo repository.RewardRepository
	logger     logger.Logger
}

// NewRewardService создаёт новый экземпляр RewardService
func NewRewardService(rewardRepo repository.RewardRepository, logger logger.Logger) RewardService {
	return &rewardService{
		rewardRepo: rewardRepo,
		logger:     logger,
	}
}

var ErrMatchAlreadyExists = errors.New("match already exists")

func (s *rewardService) RegisterReward(ctx context.Context, reward model.RewardRule) error {
	// Проверяем, существует ли правило с таким match
	exists, err := s.rewardRepo.ExistsByMatch(ctx, reward.Match)
	if err != nil {
		s.logger.Errorf("accrual: %w", err)
		return err
	}

	if exists {
		return ErrMatchAlreadyExists
	}

	// Сохраняем правило
	err = s.rewardRepo.Create(ctx, reward)
	if err != nil {
		s.logger.Errorf("accrual: %w", err)
		return err
	}

	return nil
}
