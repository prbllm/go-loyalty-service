package order

import (
	"context"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
)

type Service interface {
	Upload(ctx context.Context, userID int64, number string) error
	List(ctx context.Context, userID int64) ([]*model.Order, error)
}
