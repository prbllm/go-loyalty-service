package balance

import (
	"context"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

type service struct {
	repo   repository.Repository
	logger logger.Logger
}

func New(repo repository.Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) GetBalance(ctx context.Context, userID int64) (*model.Balance, error) {
	return s.repo.GetBalance(ctx, userID)
}

func (s *service) GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	return s.repo.GetWithdrawals(ctx, userID)
}

func (s *service) Withdraw(ctx context.Context, userID int64, orderNumber string, amount model.Amount) error {
	return s.repo.WithdrawBalance(ctx, userID, orderNumber, amount)
}
