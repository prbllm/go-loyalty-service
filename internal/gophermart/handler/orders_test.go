package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/middleware"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/service/order"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/utils"
	"go.uber.org/zap/zaptest"
)

type stubOrderService struct {
	upload func(userID int64, number string) error
	list   func(userID int64) ([]*model.Order, error)
}

func (s *stubOrderService) Upload(ctx context.Context, userID int64, number string) error {
	return s.upload(userID, number)
}

func (s *stubOrderService) List(ctx context.Context, userID int64) ([]*model.Order, error) {
	return s.list(userID)
}

func TestOrderUploadHandler(t *testing.T) {
	token, _ := utils.GenerateToken(1)

	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
	}{
		{"accepted", "79927398713", nil, http.StatusAccepted},
		{"already same user", "79927398713", order.ErrOrderAlreadyUploadedBySameUser, http.StatusOK},
		{"conflict other user", "79927398713", order.ErrOrderUploadedByAnotherUser, http.StatusConflict},
		{"invalid number", "123", order.ErrInvalidOrderNumber, http.StatusUnprocessableEntity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := zaptest.NewLogger(t).Sugar()
			handler := NewOrderHandler(&stubOrderService{
				upload: func(userID int64, number string) error {
					if userID != 1 {
						t.Fatalf("expected user id 1, got %d", userID)
					}
					return tt.serviceErr
				},
				list: nil,
			}, log)

			req := httptest.NewRequest(http.MethodPost, config.PathUserOrders, bytes.NewBufferString(tt.body))
			req.Header.Set(config.HeaderAuthorization, config.BearerPrefix+token)
			rr := httptest.NewRecorder()

			middleware.Auth(http.HandlerFunc(handler.Upload)).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestOrderUploadHandlerUnauthorized(t *testing.T) {
	log := zaptest.NewLogger(t).Sugar()
	handler := NewOrderHandler(&stubOrderService{
		upload: func(userID int64, number string) error { return nil },
		list:   nil,
	}, log)

	req := httptest.NewRequest(http.MethodPost, config.PathUserOrders, bytes.NewBufferString("79927398713"))
	rr := httptest.NewRecorder()

	handler.Upload(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestOrderListHandler(t *testing.T) {
	token, _ := utils.GenerateToken(1)
	log := zaptest.NewLogger(t).Sugar()

	now := time.Now().UTC()
	handler := NewOrderHandler(&stubOrderService{
		upload: nil,
		list: func(userID int64) ([]*model.Order, error) {
			return []*model.Order{
				{
					Number:     "111",
					Status:     model.OrderStatusProcessed,
					Accrual:    10.5,
					UploadedAt: now,
				},
				{
					Number:     "222",
					Status:     model.OrderStatusProcessing,
					Accrual:    0,
					UploadedAt: now.Add(-time.Minute),
				},
			}, nil
		},
	}, log)

	req := httptest.NewRequest(http.MethodGet, config.PathUserOrders, nil)
	req.Header.Set(config.HeaderAuthorization, config.BearerPrefix+token)
	rr := httptest.NewRecorder()

	middleware.Auth(http.HandlerFunc(handler.List)).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(resp))
	}

	if resp[0]["number"] != "111" || resp[0]["status"] != model.OrderStatusProcessed {
		t.Fatalf("unexpected first order %+v", resp[0])
	}
	if _, ok := resp[0]["accrual"]; !ok {
		t.Fatalf("expected accrual in first order")
	}
	if _, ok := resp[1]["accrual"]; ok {
		t.Fatalf("did not expect accrual for non processed order")
	}
}

func TestOrderListHandlerEmpty(t *testing.T) {
	token, _ := utils.GenerateToken(1)
	log := zaptest.NewLogger(t).Sugar()

	handler := NewOrderHandler(&stubOrderService{
		upload: nil,
		list: func(userID int64) ([]*model.Order, error) {
			return []*model.Order{}, nil
		},
	}, log)

	req := httptest.NewRequest(http.MethodGet, config.PathUserOrders, nil)
	req.Header.Set(config.HeaderAuthorization, config.BearerPrefix+token)
	rr := httptest.NewRecorder()

	middleware.Auth(http.HandlerFunc(handler.List)).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}
