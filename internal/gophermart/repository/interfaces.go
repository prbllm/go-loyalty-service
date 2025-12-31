package repository

import (
	"context"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
)

//go:generate mockgen -source=interfaces.go -destination=../../mocks/gophermart/repository.go -package=mocks

type Repository interface {
	CreateUser(ctx context.Context, login string, passwordHash string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	Close() error

	CreateOrder(ctx context.Context, userID int64, orderNumber string) error
	GetOrderByNumber(ctx context.Context, orderNumber string) (*model.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64) ([]*model.Order, error)
	GetOrdersByStatus(ctx context.Context, status model.OrderStatus) ([]*model.Order, error)
	UpdateOrderStatus(ctx context.Context, orderNumber string, status model.OrderStatus, accrual model.Amount) error

	GetBalance(ctx context.Context, userID int64) (*model.Balance, error)
	WithdrawBalance(ctx context.Context, userID int64, orderNumber string, amount model.Amount) error
	GetWithdrawals(ctx context.Context, userID int64) ([]*model.Withdrawal, error)
	AddAccrual(ctx context.Context, userID int64, amount model.Amount) error
}
