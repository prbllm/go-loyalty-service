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

func TestWorkerPool_FetchAndQueueOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)

	orderNew := &model.Order{Number: "1", Status: model.OrderStatusNew}
	orderProcessing := &model.Order{Number: "2", Status: model.OrderStatusProcessing}

	repo.EXPECT().GetOrdersByStatus(gomock.Any(), model.OrderStatusNew).Return([]*model.Order{orderNew}, nil)
	repo.EXPECT().GetOrdersByStatus(gomock.Any(), model.OrderStatusProcessing).Return([]*model.Order{orderProcessing}, nil)

	client := clientFunc(func(ctx context.Context, number string) (*Response, error) {
		return nil, ErrOrderNotRegistered
	})

	pool := NewWorkerPool(repo, client, logger.NewNop(), time.Second, 2)
	pool.jobs = make(chan *model.Order, 10)

	backoff := pool.fetchAndQueueOrders(context.Background())
	if backoff != 0 {
		t.Fatalf("expected no backoff, got %s", backoff)
	}

	select {
	case order := <-pool.jobs:
		if order.Number != "1" && order.Number != "2" {
			t.Fatalf("unexpected order number: %s", order.Number)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected order in channel")
	}
}

func TestWorkerPool_HandleOrder_Processed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)

	order := &model.Order{Number: "1"}

	repo.EXPECT().UpdateOrderStatus(gomock.Any(), order.Number, model.OrderStatusProcessed, model.Amount(1250)).Return(nil)

	client := clientFunc(func(ctx context.Context, number string) (*Response, error) {
		return &Response{Order: number, Status: StatusProcessed, Accrual: 12.5}, nil
	})

	pool := NewWorkerPool(repo, client, logger.NewNop(), time.Second, 1)
	pool.handleOrder(context.Background(), order, 0)
}

func TestWorkerPool_HandleOrder_Invalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)

	order := &model.Order{Number: "2"}

	repo.EXPECT().UpdateOrderStatus(gomock.Any(), order.Number, model.OrderStatusInvalid, model.Amount(0)).Return(nil)

	client := clientFunc(func(ctx context.Context, number string) (*Response, error) {
		return &Response{Order: number, Status: StatusInvalid}, nil
	})

	pool := NewWorkerPool(repo, client, logger.NewNop(), time.Second, 1)
	pool.handleOrder(context.Background(), order, 0)
}

func TestWorkerPool_HandleOrder_TooManyRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)

	order := &model.Order{Number: "3"}

	client := clientFunc(func(ctx context.Context, number string) (*Response, error) {
		return nil, &TooManyRequestsError{RetryAfter: 2 * time.Second}
	})

	pool := NewWorkerPool(repo, client, logger.NewNop(), time.Second, 1)

	pool.handleOrder(context.Background(), order, 0)

	select {
	case retryAfter := <-pool.rateLimitChan:
		if retryAfter != 2*time.Second {
			t.Fatalf("expected retry after 2s, got %s", retryAfter)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected rate limit signal")
	}
}

func TestWorkerPool_HandleOrder_OrderNotRegistered(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)

	order := &model.Order{Number: "4"}

	client := clientFunc(func(ctx context.Context, number string) (*Response, error) {
		return nil, ErrOrderNotRegistered
	})

	pool := NewWorkerPool(repo, client, logger.NewNop(), time.Second, 1)
	pool.handleOrder(context.Background(), order, 0)
}

func TestWorkerPool_HandleOrder_UnknownError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)

	order := &model.Order{Number: "5"}

	client := clientFunc(func(ctx context.Context, number string) (*Response, error) {
		return nil, errors.New("some error")
	})

	pool := NewWorkerPool(repo, client, logger.NewNop(), time.Second, 1)
	pool.handleOrder(context.Background(), order, 0)
}

func TestMapStatus(t *testing.T) {
	tests := []struct {
		name          string
		accrualStatus string
		accrual       float64
		wantStatus    model.OrderStatus
		wantAccrual   model.Amount
	}{
		{"registered", StatusRegistered, 0, model.OrderStatusProcessing, model.Amount(0)},
		{"processing", StatusProcessing, 0, model.OrderStatusProcessing, model.Amount(0)},
		{"invalid", StatusInvalid, 0, model.OrderStatusInvalid, model.Amount(0)},
		{"processed", StatusProcessed, 12.5, model.OrderStatusProcessed, model.FromFloat64(12.5)},
		{"unknown", "UNKNOWN", 0, "", model.Amount(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, accrual := mapStatus(tt.accrualStatus, tt.accrual)
			if status != tt.wantStatus {
				t.Errorf("mapStatus() status = %v, want %v", status, tt.wantStatus)
			}
			if accrual != tt.wantAccrual {
				t.Errorf("mapStatus() accrual = %v, want %v", accrual, tt.wantAccrual)
			}
		})
	}
}
