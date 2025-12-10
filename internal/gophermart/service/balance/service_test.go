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

	mockRepo.EXPECT().GetBalance(gomock.Any(), int64(1)).Return(&model.Balance{Current: model.Amount(1000), Withdrawn: model.Amount(200)}, nil)
	mockRepo.EXPECT().GetWithdrawals(gomock.Any(), int64(1)).Return([]*model.Withdrawal{
		{OrderNumber: "1", Sum: model.Amount(200)},
	}, nil)

	bal, err := svc.GetBalance(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bal.Current != model.Amount(1000) || bal.Withdrawn != model.Amount(200) {
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

	mockRepo.EXPECT().WithdrawBalance(gomock.Any(), int64(1), "order", model.Amount(500)).Return(nil)

	if err := svc.Withdraw(context.Background(), 1, "order", model.Amount(500)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(mockRepo, log)

	mockRepo.EXPECT().WithdrawBalance(gomock.Any(), int64(1), "order", model.Amount(500)).Return(repository.ErrInsufficientFunds)

	err := svc.Withdraw(context.Background(), 1, "order", model.Amount(500))
	if !errors.Is(err, repository.ErrInsufficientFunds) {
		t.Fatalf("expected insufficient funds, got %v", err)
	}
}
