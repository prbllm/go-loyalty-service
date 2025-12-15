package repository

import (
	"context"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
)

// OrderRepository отвечает за операции с заказами
type OrderRepository interface {
	// Create создаёт новый заказ со статусом REGISTERED
	Create(ctx context.Context, order *model.Order) error

	// GetByNumber возвращает заказ по номеру, если существует
	GetByNumber(ctx context.Context, number string) (*model.Order, error)

	// UpdateStatusAndAccrual обновляет статус и сумму начисления для заказа
	UpdateStatusAndAccrual(ctx context.Context, number string, status model.OrderStatus, accrual *int64) error

	// GetAllProcessing возвращает все заказы со статусом PROCESSING (для фоновой обработки)
	GetAllProcessing(ctx context.Context) ([]model.Order, error)
}
