package service

import (
	"context"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/accrual/repository"
)

// OrderService отвечает за бизнес-логику, связанную с заказами
type OrderService interface {
	RegisterOrder(ctx context.Context, number string, goods []model.Good) error
	GetOrder(ctx context.Context, number string) (*model.Order, error)
	ProcessOrder(ctx context.Context, order *model.Order) (*int64, error) // возвращает accrual
}

// orderService — реализация OrderService
type orderService struct {
	orderRepo  repository.OrderRepository
	rewardRepo repository.RewardRepository
}

// NewOrderService создаёт новый экземпляр OrderService
func NewOrderService(
	orderRepo repository.OrderRepository,
	rewardRepo repository.RewardRepository,
) OrderService {
	return &orderService{
		orderRepo:  orderRepo,
		rewardRepo: rewardRepo,
	}
}

func (s *orderService) RegisterOrder(ctx context.Context, number string, goods []model.Good) error {
	panic("not implemented")
}

func (s *orderService) GetOrder(ctx context.Context, number string) (*model.Order, error) {
	panic("not implemented")
}

func (s *orderService) ProcessOrder(ctx context.Context, order *model.Order) (*int64, error) {
	panic("not implemented")
}
