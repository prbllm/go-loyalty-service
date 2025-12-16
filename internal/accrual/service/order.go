package service

import (
	"context"
	"errors"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/accrual/repository"
)

// OrderService отвечает за бизнес-логику, связанную с заказами
type OrderService interface {
	RegisterOrder(ctx context.Context, order model.Order) error
	GetOrder(ctx context.Context, number string) (model.Order, error)
	ProcessOrder(ctx context.Context, order *model.Order) (*int64, error) // возвращает accrual
}

// orderService — реализация OrderService
type orderService struct {
	orderRepo  repository.OrderRepository
	rewardRepo repository.RewardRepository
}

// NewOrderService создаёт новый экземпляр OrderService
func NewOrderService(orderRepo repository.OrderRepository, rewardRepo repository.RewardRepository) OrderService {
	return &orderService{
		orderRepo:  orderRepo,
		rewardRepo: rewardRepo,
	}
}

var ErrOrderAlreadyExists = errors.New("order already exists")

func (s *orderService) RegisterOrder(ctx context.Context, order model.Order) error {
	// Проверяем, существует ли заказ с таким номером
	exists, err := s.orderRepo.IsOrderExists(ctx, order.Number)
	// Другая ошибка БД
	if err != nil {
		return err
	}
	// Заказ найден → дубликат
	if exists {
		return ErrOrderAlreadyExists
	}

	err = s.orderRepo.Create(ctx, order)

	if err != nil {
		return err
	}

	return nil
}

var ErrOrderNotFound = errors.New("order not found")

func (s *orderService) GetOrder(ctx context.Context, number string) (model.Order, error) {
	var order model.Order
	// Проверяем, существует ли заказ с таким номером
	exists, err := s.orderRepo.IsOrderExists(ctx, number)
	// Другая ошибка БД
	if err != nil {
		return order, err
	}
	// Заказ не найден
	if !exists {
		return order, ErrOrderNotFound
	}

	order, err = s.orderRepo.GetByNumber(ctx, number)

	if err != nil {
		return order, err
	}

	return order, nil
}

func (s *orderService) ProcessOrder(ctx context.Context, order *model.Order) (*int64, error) {
	panic("not implemented")
}
