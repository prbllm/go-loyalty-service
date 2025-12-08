package auth

import "context"

//go:generate mockgen -source=interfaces.go -destination=../../mocks/gophermart/auth_service.go -package=mocks

type Service interface {
	Register(ctx context.Context, login string, password string) (string, error)
	Login(ctx context.Context, login string, password string) (string, error)
}
