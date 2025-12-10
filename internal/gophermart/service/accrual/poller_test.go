package accrual

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/logger"
	mocks "github.com/prbllm/go-loyalty-service/internal/mocks/gophermart"
	"go.uber.org/mock/gomock"
)

type clientFunc func(ctx context.Context, number string) (*Response, error)

func (f clientFunc) GetOrder(ctx context.Context, number string) (*Response, error) {
	return f(ctx, number)
}

func TestPollerProcess_UpdatesStatuses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)

	orderNew := &model.Order{Number: "1"}

	repo.EXPECT().GetOrdersByStatus(gomock.Any(), model.OrderStatusNew).Return([]*model.Order{orderNew}, nil)
	repo.EXPECT().GetOrdersByStatus(gomock.Any(), model.OrderStatusProcessing).Return(nil, nil)
	repo.EXPECT().UpdateOrderStatus(gomock.Any(), orderNew.Number, model.OrderStatusProcessed, model.Amount(1250)).Return(nil)

	client := clientFunc(func(ctx context.Context, number string) (*Response, error) {
		return &Response{Order: number, Status: StatusProcessed, Accrual: 12.5}, nil
	})

	poller := NewPoller(repo, client, logger.NewNop(), time.Second)

	if backoff := poller.process(context.Background()); backoff != 0 {
		t.Fatalf("expected no backoff, got %s", backoff)
	}
}

func TestPollerProcess_Invalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	orderProcessing := &model.Order{Number: "2"}

	repo.EXPECT().GetOrdersByStatus(gomock.Any(), model.OrderStatusNew).Return(nil, nil)
	repo.EXPECT().GetOrdersByStatus(gomock.Any(), model.OrderStatusProcessing).Return([]*model.Order{orderProcessing}, nil)
	repo.EXPECT().UpdateOrderStatus(gomock.Any(), orderProcessing.Number, model.OrderStatusInvalid, model.Amount(0)).Return(nil)

	client := clientFunc(func(ctx context.Context, number string) (*Response, error) {
		return &Response{Order: number, Status: StatusInvalid}, nil
	})

	poller := NewPoller(repo, client, logger.NewNop(), time.Second)

	if backoff := poller.process(context.Background()); backoff != 0 {
		t.Fatalf("expected no backoff, got %s", backoff)
	}
}

func TestPollerProcess_TooManyRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	orderNew := &model.Order{Number: "3"}

	repo.EXPECT().GetOrdersByStatus(gomock.Any(), model.OrderStatusNew).Return([]*model.Order{orderNew}, nil)
	repo.EXPECT().GetOrdersByStatus(gomock.Any(), model.OrderStatusProcessing).Return(nil, nil)

	client := clientFunc(func(ctx context.Context, number string) (*Response, error) {
		return nil, &TooManyRequestsError{RetryAfter: 2 * time.Second}
	})

	poller := NewPoller(repo, client, logger.NewNop(), time.Second)

	if backoff := poller.process(context.Background()); backoff != 2*time.Second {
		t.Fatalf("expected backoff 2s, got %s", backoff)
	}
}

func TestPollerProcess_UnknownErrorLogged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	orderNew := &model.Order{Number: "4"}

	repo.EXPECT().GetOrdersByStatus(gomock.Any(), model.OrderStatusNew).Return([]*model.Order{orderNew}, nil)
	repo.EXPECT().GetOrdersByStatus(gomock.Any(), model.OrderStatusProcessing).Return(nil, nil)

	client := clientFunc(func(ctx context.Context, number string) (*Response, error) {
		return nil, errors.New("some error")
	})

	poller := NewPoller(repo, client, logger.NewNop(), time.Second)

	if backoff := poller.process(context.Background()); backoff != 0 {
		t.Fatalf("expected no backoff, got %s", backoff)
	}
}
