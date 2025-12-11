package balance

import (
	"context"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
)

//go:generate mockgen -source=interfaces.go -destination=../../../mocks/gophermart/balance_service.go -package=mocks -mock_names Service=MockBalanceService

type Service interface {
	GetBalance(ctx context.Context, userID int64) (*model.Balance, error)
	GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error)
	Withdraw(ctx context.Context, userID int64, orderNumber string, amount model.Amount) error
}
