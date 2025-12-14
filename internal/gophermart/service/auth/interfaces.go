package auth

import "context"

//go:generate mockgen -source=interfaces.go -destination=../../../mocks/gophermart/auth_service.go -package=mocks -mock_names Service=MockAuthService

type Service interface {
	Register(ctx context.Context, login string, password string) (string, error)
	Login(ctx context.Context, login string, password string) (string, error)
}
