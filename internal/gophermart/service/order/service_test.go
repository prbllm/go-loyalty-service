package order

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	mocks "github.com/prbllm/go-loyalty-service/internal/mocks/gophermart"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestUploadNewOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(repo, log)

	repo.EXPECT().GetOrderByNumber(gomock.Any(), "79927398713").Return(nil, sql.ErrNoRows)
	repo.EXPECT().CreateOrder(gomock.Any(), int64(1), "79927398713").Return(nil)

	if err := svc.Upload(context.Background(), 1, "79927398713"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUploadInvalidNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(repo, log)

	err := svc.Upload(context.Background(), 1, "123")
	if !errors.Is(err, ErrInvalidOrderNumber) {
		t.Fatalf("expected invalid number error, got %v", err)
	}
}

func TestUploadExistingSameUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(repo, log)

	repo.EXPECT().GetOrderByNumber(gomock.Any(), "79927398713").Return(&model.Order{
		UserID: 1,
	}, nil)

	err := svc.Upload(context.Background(), 1, "79927398713")
	if !errors.Is(err, ErrOrderAlreadyUploadedBySameUser) {
		t.Fatalf("expected same user error, got %v", err)
	}
}

func TestUploadExistingOtherUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(repo, log)

	repo.EXPECT().GetOrderByNumber(gomock.Any(), "79927398713").Return(&model.Order{
		UserID: 2,
	}, nil)

	err := svc.Upload(context.Background(), 1, "79927398713")
	if !errors.Is(err, ErrOrderUploadedByAnotherUser) {
		t.Fatalf("expected other user error, got %v", err)
	}
}

func TestUploadRaceDuplicateSameUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(repo, log)

	gomock.InOrder(
		repo.EXPECT().GetOrderByNumber(gomock.Any(), "79927398713").Return(nil, sql.ErrNoRows),
		repo.EXPECT().CreateOrder(gomock.Any(), int64(1), "79927398713").Return(repository.ErrOrderAlreadyExists),
		repo.EXPECT().GetOrderByNumber(gomock.Any(), "79927398713").Return(&model.Order{
			UserID: 1,
		}, nil),
	)

	err := svc.Upload(context.Background(), 1, "79927398713")
	if !errors.Is(err, ErrOrderAlreadyUploadedBySameUser) {
		t.Fatalf("expected same user error, got %v", err)
	}
}

func TestUploadRaceDuplicateOtherUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(repo, log)

	gomock.InOrder(
		repo.EXPECT().GetOrderByNumber(gomock.Any(), "79927398713").Return(nil, sql.ErrNoRows),
		repo.EXPECT().CreateOrder(gomock.Any(), int64(1), "79927398713").Return(repository.ErrOrderAlreadyExists),
		repo.EXPECT().GetOrderByNumber(gomock.Any(), "79927398713").Return(&model.Order{
			UserID: 2,
		}, nil),
	)

	err := svc.Upload(context.Background(), 1, "79927398713")
	if !errors.Is(err, ErrOrderUploadedByAnotherUser) {
		t.Fatalf("expected other user error, got %v", err)
	}
}

func TestListOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(repo, log)

	orders := []*model.Order{
		{ID: 2, UserID: 1},
		{ID: 1, UserID: 1},
	}

	repo.EXPECT().GetOrdersByUserID(gomock.Any(), int64(1)).Return(orders, nil)

	result, err := svc.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != len(orders) || result[0].ID != 2 || result[1].ID != 1 {
		t.Fatalf("unexpected orders result %+v", result)
	}
}
