package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/service/auth"
	"go.uber.org/zap/zaptest"
)

type stubAuthService struct {
	register func(login, password string) (string, error)
	login    func(login, password string) (string, error)
}

func (s *stubAuthService) Register(ctx context.Context, login string, password string) (string, error) {
	return s.register(login, password)
}

func (s *stubAuthService) Login(ctx context.Context, login string, password string) (string, error) {
	return s.login(login, password)
}

func TestRegisterHandler(t *testing.T) {
	log := zaptest.NewLogger(t).Sugar()
	handler := NewAuthHandler(&stubAuthService{
		register: func(login, password string) (string, error) {
			if login == "exists" {
				return "", repository.ErrUserAlreadyExists
			}
			return "token123", nil
		},
		login: nil,
	}, log)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(`{"login":"user","password":"pass"}`))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
		if got := rr.Header().Get("Authorization"); got != "Bearer token123" {
			t.Fatalf("expected auth header, got %s", got)
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(`{"login":"exists","password":"pass"}`))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rr.Code)
		}
	})

	t.Run("bad body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(`{`))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})
}

func TestLoginHandler(t *testing.T) {
	log := zaptest.NewLogger(t).Sugar()
	handler := NewAuthHandler(&stubAuthService{
		login: func(login, password string) (string, error) {
			if login == "bad" {
				return "", auth.ErrInvalidCredentials
			}
			return "token456", nil
		},
		register: nil,
	}, log)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(`{"login":"user","password":"pass"}`))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
		if got := rr.Header().Get("Authorization"); got != "Bearer token456" {
			t.Fatalf("expected auth header, got %s", got)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(`{"login":"bad","password":"pass"}`))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("bad body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(`{`))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})
}
