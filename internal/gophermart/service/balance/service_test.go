package balance

import (
	"context"
	"errors"
	"testing"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	mocks "github.com/prbllm/go-loyalty-service/internal/mocks/gophermart"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestGetBalanceAndWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(mockRepo, log)

	mockRepo.EXPECT().GetBalance(gomock.Any(), int64(1)).Return(&model.Balance{Current: 10, Withdrawn: 2}, nil)
	mockRepo.EXPECT().GetWithdrawals(gomock.Any(), int64(1)).Return([]*model.Withdrawal{
		{OrderNumber: "1", Sum: 2},
	}, nil)

	bal, err := svc.GetBalance(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bal.Current != 10 || bal.Withdrawn != 2 {
		t.Fatalf("unexpected balance: %+v", bal)
	}

	list, err := svc.GetWithdrawals(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 || list[0].OrderNumber != "1" {
		t.Fatalf("unexpected withdrawals: %+v", list)
	}
}

func TestWithdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(mockRepo, log)

	mockRepo.EXPECT().WithdrawBalance(gomock.Any(), int64(1), "order", 5.0).Return(nil)

	if err := svc.Withdraw(context.Background(), 1, "order", 5.0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(mockRepo, log)

	mockRepo.EXPECT().WithdrawBalance(gomock.Any(), int64(1), "order", 5.0).Return(repository.ErrInsufficientFunds)

	err := svc.Withdraw(context.Background(), 1, "order", 5.0)
	if !errors.Is(err, repository.ErrInsufficientFunds) {
		t.Fatalf("expected insufficient funds, got %v", err)
	}
}
