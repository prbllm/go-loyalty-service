package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/middleware"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/utils"
	balancemocks "github.com/prbllm/go-loyalty-service/internal/mocks/gophermart"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func addAuthHeader(req *http.Request, userID int64) {
	token, _ := utils.GenerateToken(userID)
	req.Header.Set(config.HeaderAuthorization, config.BearerPrefix+token)
}

func TestBalanceHandler_Balance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := balancemocks.NewMockBalanceService(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	h := NewBalanceHandler(mockService, log)

	mockService.EXPECT().GetBalance(gomock.Any(), int64(1)).Return(&model.Balance{Current: 10, Withdrawn: 2}, nil)

	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	addAuthHeader(req, 1)
	rr := httptest.NewRecorder()

	middleware.Auth(http.HandlerFunc(h.Balance)).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestBalanceHandler_Withdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := balancemocks.NewMockBalanceService(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	h := NewBalanceHandler(mockService, log)

	now := time.Now()
	mockService.EXPECT().GetWithdrawals(gomock.Any(), int64(1)).Return([]*model.Withdrawal{
		{OrderNumber: "1", Sum: 5, ProcessedAt: now},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
	addAuthHeader(req, 1)
	rr := httptest.NewRecorder()

	middleware.Auth(http.HandlerFunc(h.Withdrawals)).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestBalanceHandler_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := balancemocks.NewMockBalanceService(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	h := NewBalanceHandler(mockService, log)

	mockService.EXPECT().Withdraw(gomock.Any(), int64(1), "79927398713", 5.0).Return(nil)

	body := bytes.NewBufferString(`{"order":"79927398713","sum":5}`)
	req := httptest.NewRequest(http.MethodPost, "/withdraw", body)
	addAuthHeader(req, 1)
	rr := httptest.NewRecorder()

	middleware.Auth(http.HandlerFunc(h.Withdraw)).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestBalanceHandler_Withdraw_Insufficient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := balancemocks.NewMockBalanceService(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	h := NewBalanceHandler(mockService, log)

	mockService.EXPECT().Withdraw(gomock.Any(), int64(1), "79927398713", 5.0).Return(repository.ErrInsufficientFunds)

	body := bytes.NewBufferString(`{"order":"79927398713","sum":5}`)
	req := httptest.NewRequest(http.MethodPost, "/withdraw", body)
	addAuthHeader(req, 1)
	rr := httptest.NewRecorder()

	middleware.Auth(http.HandlerFunc(h.Withdraw)).ServeHTTP(rr, req)

	if rr.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d", rr.Code)
	}
}

func TestBalanceHandler_Withdraw_InvalidOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := balancemocks.NewMockBalanceService(ctrl)
	log := zaptest.NewLogger(t).Sugar()
	h := NewBalanceHandler(mockService, log)

	body := bytes.NewBufferString(`{"order":"abc","sum":5}`)
	req := httptest.NewRequest(http.MethodPost, "/withdraw", body)
	addAuthHeader(req, 1)
	rr := httptest.NewRecorder()

	middleware.Auth(http.HandlerFunc(h.Withdraw)).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rr.Code)
	}
}
