package service

import (
	"context"
	"errors"
	"strings"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/accrual/repository"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

//go:generate mockgen -source=order.go -destination=../../mocks/accrual/order_service.go -package=mocks

// OrderService отвечает за бизнес-логику, связанную с заказами
type OrderService interface {
	RegisterOrder(ctx context.Context, order model.Order) error
	GetOrder(ctx context.Context, number string) (*model.Order, error)
}

// orderService — реализация OrderService
type orderService struct {
	orderRepo  repository.OrderRepository
	rewardRepo repository.RewardRepository
	logger     logger.Logger
}

// NewOrderService создаёт новый экземпляр OrderService
func NewOrderService(orderRepo repository.OrderRepository, rewardRepo repository.RewardRepository, logger logger.Logger) OrderService {
	return &orderService{
		orderRepo:  orderRepo,
		rewardRepo: rewardRepo,
		logger:     logger,
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

	go func() {
		ctx := context.Background()
		s.setOrderProcessing(ctx, order.Number)
		accrual, err := s.processOrder(ctx, &order)
		if err != nil {
			s.logger.Error(err)
			s.setOrderInvalid(ctx, order.Number)
		} else {
			s.setOrderProcessed(ctx, order.Number, &accrual)
		}
	}()

	return nil
}

var ErrOrderNotFound = errors.New("order not found")

func (s *orderService) GetOrder(ctx context.Context, number string) (*model.Order, error) {
	// Проверяем, существует ли заказ с таким номером
	exists, err := s.orderRepo.IsOrderExists(ctx, number)
	// Другая ошибка БД
	if err != nil {
		return nil, err
	}
	// Заказ не найден
	if !exists {
		return nil, ErrOrderNotFound
	}

	order, err := s.orderRepo.GetByNumber(ctx, number)

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *orderService) processOrder(ctx context.Context, order *model.Order) (int64, error) {
	// Получаем все правила начисления
	rules, err := s.rewardRepo.GetAll(ctx)
	if err != nil {
		return 0, err
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

	// Возвращаем в копейках
	return int64(totalAccrualRub * 100), nil

}

func (s *orderService) setOrderProcessing(ctx context.Context, number string) error {
	return s.orderRepo.UpdateStatusAndAccrual(ctx, number, model.Processing, nil)
}

func (s *orderService) setOrderInvalid(ctx context.Context, number string) error {
	return s.orderRepo.UpdateStatusAndAccrual(ctx, number, model.Invalid, nil)
}

func (s *orderService) setOrderProcessed(ctx context.Context, number string, accrual *int64) error {
	return s.orderRepo.UpdateStatusAndAccrual(ctx, number, model.Processed, accrual)
}
