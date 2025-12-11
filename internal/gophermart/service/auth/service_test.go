package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/utils"
	mocks "github.com/prbllm/go-loyalty-service/internal/mocks/gophermart"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()

	svc := New(mockRepo, log)

	mockRepo.EXPECT().CreateUser(gomock.Any(), "login", gomock.Any()).Return(int64(1), nil)

	token, err := svc.Register(context.Background(), "login", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userID, err := utils.ParseToken(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if userID != 1 {
		t.Fatalf("expected user id 1, got %d", userID)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(mockRepo, log)

	mockRepo.EXPECT().CreateUser(gomock.Any(), "login", gomock.Any()).Return(int64(0), repository.ErrUserAlreadyExists)

	_, err := svc.Register(context.Background(), "login", "password")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, repository.ErrUserAlreadyExists) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(mockRepo, log)

	hash, _ := utils.HashPassword("password")
	mockRepo.EXPECT().GetUserByLogin(gomock.Any(), "login").Return(&model.User{
		ID:           2,
		Login:        "login",
		PasswordHash: hash,
	}, nil)

	token, err := svc.Login(context.Background(), "login", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userID, err := utils.ParseToken(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if userID != 2 {
		t.Fatalf("expected user id 2, got %d", userID)
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(mockRepo, log)

	hash, _ := utils.HashPassword("password")
	mockRepo.EXPECT().GetUserByLogin(gomock.Any(), "login").Return(&model.User{
		ID:           2,
		Login:        "login",
		PasswordHash: hash,
	}, nil)

	_, err := svc.Login(context.Background(), "login", "wrong")
	if err == nil || !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}

func TestLoginNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	svc := New(mockRepo, log)

	mockRepo.EXPECT().GetUserByLogin(gomock.Any(), "login").Return(nil, sql.ErrNoRows)

	_, err := svc.Login(context.Background(), "login", "password")
	if err == nil || !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}
