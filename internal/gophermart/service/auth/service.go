package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/utils"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type service struct {
	repo   repository.Repository
	logger logger.Logger
}

func New(repo repository.Repository, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) Register(ctx context.Context, login string, password string) (string, error) {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	userID, err := s.repo.CreateUser(ctx, login, hash)
	if err != nil {
		return "", err
	}

	token, err := utils.GenerateToken(userID)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

func (s *service) Login(ctx context.Context, login string, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if !utils.CheckPassword(user.PasswordHash, password) {
		return "", ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}
