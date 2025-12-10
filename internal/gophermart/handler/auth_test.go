package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/service/auth"
	authmocks "github.com/prbllm/go-loyalty-service/internal/mocks/gophermart"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestRegisterHandler(t *testing.T) {
	log := zaptest.NewLogger(t).Sugar()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := authmocks.NewMockAuthService(ctrl)
		handler := NewAuthHandler(mockService, log)

		mockService.EXPECT().Register(gomock.Any(), "user", "pass").Return("token123", nil)

		req := httptest.NewRequest(http.MethodPost, config.PathUserRegister, bytes.NewBufferString(`{"login":"user","password":"pass"}`))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
		if got := rr.Header().Get(config.HeaderAuthorization); got != config.BearerPrefix+"token123" {
			t.Fatalf("expected auth header, got %s", got)
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := authmocks.NewMockAuthService(ctrl)
		handler := NewAuthHandler(mockService, log)

		mockService.EXPECT().Register(gomock.Any(), "exists", "pass").Return("", repository.ErrUserAlreadyExists)

		req := httptest.NewRequest(http.MethodPost, config.PathUserRegister, bytes.NewBufferString(`{"login":"exists","password":"pass"}`))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rr.Code)
		}
	})

	t.Run("bad body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := authmocks.NewMockAuthService(ctrl)
		handler := NewAuthHandler(mockService, log)

		req := httptest.NewRequest(http.MethodPost, config.PathUserRegister, bytes.NewBufferString(`{`))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})
}

func TestLoginHandler(t *testing.T) {
	log := zaptest.NewLogger(t).Sugar()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := authmocks.NewMockAuthService(ctrl)
		handler := NewAuthHandler(mockService, log)

		mockService.EXPECT().Login(gomock.Any(), "user", "pass").Return("token456", nil)

		req := httptest.NewRequest(http.MethodPost, config.PathUserLogin, bytes.NewBufferString(`{"login":"user","password":"pass"}`))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
		if got := rr.Header().Get(config.HeaderAuthorization); got != config.BearerPrefix+"token456" {
			t.Fatalf("expected auth header, got %s", got)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := authmocks.NewMockAuthService(ctrl)
		handler := NewAuthHandler(mockService, log)

		mockService.EXPECT().Login(gomock.Any(), "bad", "pass").Return("", auth.ErrInvalidCredentials)

		req := httptest.NewRequest(http.MethodPost, config.PathUserLogin, bytes.NewBufferString(`{"login":"bad","password":"pass"}`))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("bad body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := authmocks.NewMockAuthService(ctrl)
		handler := NewAuthHandler(mockService, log)

		req := httptest.NewRequest(http.MethodPost, config.PathUserLogin, bytes.NewBufferString(`{`))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})
}
