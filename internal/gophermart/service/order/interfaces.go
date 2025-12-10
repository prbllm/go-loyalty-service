package order

import (
	"context"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
)

//go:generate mockgen -source=interfaces.go -destination=../../../mocks/gophermart/order_service.go -package=mocks -mock_names Service=MockOrderService

type Service interface {
	Upload(ctx context.Context, userID int64, number string) error
	List(ctx context.Context, userID int64) ([]*model.Order, error)
}
