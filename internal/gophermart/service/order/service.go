package order

import (
	"context"
	"database/sql"
	"errors"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/utils"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

var (
	ErrInvalidOrderNumber             = errors.New("invalid order number")
	ErrOrderAlreadyUploadedBySameUser = errors.New("order already uploaded by the same user")
	ErrOrderUploadedByAnotherUser     = errors.New("order uploaded by another user")
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

func (s *service) Upload(ctx context.Context, userID int64, number string) error {
	if !utils.IsValidOrderNumber(number) {
		return ErrInvalidOrderNumber
	}

	order, err := s.repo.GetOrderByNumber(ctx, number)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := s.repo.CreateOrder(ctx, userID, number); err != nil {
				if errors.Is(err, repository.ErrOrderAlreadyExists) {
					return s.resolveExistingOrder(ctx, userID, number)
				}
				return err
			}
			return nil
		}
		return err
	}

	if order.UserID == userID {
		return ErrOrderAlreadyUploadedBySameUser
	}

	return ErrOrderUploadedByAnotherUser
}

func (s *service) resolveExistingOrder(ctx context.Context, userID int64, number string) error {
	order, err := s.repo.GetOrderByNumber(ctx, number)
	if err != nil {
		return err
	}

	if order.UserID == userID {
		return ErrOrderAlreadyUploadedBySameUser
	}

	return ErrOrderUploadedByAnotherUser
}

func (s *service) List(ctx context.Context, userID int64) ([]*model.Order, error) {
	return s.repo.GetOrdersByUserID(ctx, userID)
}
