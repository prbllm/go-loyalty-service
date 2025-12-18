package service

import (
	"context"
	"errors"
	"strings"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/accrual/repository"
)

//go:generate mockgen -source=order.go -destination=../../mocks/accrual/order_service.go -package=mocks

// OrderService отвечает за бизнес-логику, связанную с заказами
type OrderService interface {
	RegisterOrder(ctx context.Context, order model.Order) error
	GetOrder(ctx context.Context, number string) (model.Order, error)
	ProcessOrder(ctx context.Context, order *model.Order) (*int64, error) // возвращает accrual
	SetOrderProcessing(ctx context.Context, number string) error
	SetOrderProcessed(ctx context.Context, number string, accrual *int64) error
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

	s.SetOrderProcessing(ctx, order.Number)
	accrual, _ := s.ProcessOrder(ctx, &order)
	s.SetOrderProcessed(ctx, order.Number, accrual)

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
	// Получаем все правила начисления
	rules, err := s.rewardRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var totalAccrualRub float64 // накапливаем в рублях (дробно)

	// Проходим по каждому товару в заказе
	for _, good := range order.Goods {
		// Ищем первое правило, match которого содержится в описании товара
		for _, rule := range rules {
			if strings.Contains(good.Description, rule.Match) {
				var accrualRub float64
				switch rule.RewardType {
				case model.RewardTypePercent:
					// Начисление = цена * процент / 100
					accrualRub = (float64(good.Price)*rule.Reward + 50) / 100.00 / 100.00
				case model.RewardTypePoints:
					// reward — уже в баллах (рублях), может быть дробным
					accrualRub = rule.Reward
				}
				totalAccrualRub += accrualRub
				break // одно правило на товар
			}
		}
	}

	finalAccrual := int64(totalAccrualRub * 100)

	// Если итог <= 0 — возвращаем nil (поле accrual отсутствует в JSON)
	if finalAccrual <= 0 {
		return nil, nil
	}

	return &finalAccrual, nil

}

func (s *orderService) SetOrderProcessing(ctx context.Context, number string) error {
	return s.orderRepo.UpdateStatusAndAccrual(ctx, number, model.Processing, nil)
}

func (s *orderService) SetOrderProcessed(ctx context.Context, number string, accrual *int64) error {
	return s.orderRepo.UpdateStatusAndAccrual(ctx, number, model.Processed, accrual)
}
