package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prbllm/go-loyalty-service/internal/accrual/handler"
	"github.com/prbllm/go-loyalty-service/internal/accrual/service/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_GetOrderInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mock.NewMockOrderService(ctrl)
	mockReward := mock.NewMockRewardService(ctrl)
	h := handler.New(mockOrder, mockReward)

	tests := []struct {
		name                string
		url                 string
		expectedStatus      int
		expectedContentType string
	}{
		{
			name:                "valid order number",
			url:                 "/api/orders/1234567890",
			expectedStatus:      http.StatusOK,
			expectedContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			h.GetOrderInfo(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
			require.Equal(t, tt.expectedContentType, w.Header().Get("Content-Type"))
		})
	}
}

func TestHandler_RegisterOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mock.NewMockOrderService(ctrl)
	mockReward := mock.NewMockRewardService(ctrl)
	h := handler.New(mockOrder, mockReward)

	tests := []struct {
		name           string
		contentType    string
		body           string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid JSON",
			contentType:    "application/json",
			body:           `{"order": "1234567890"}`,
			expectedStatus: http.StatusAccepted,
			expectError:    false,
		},
		{
			name:           "invalid content-type",
			contentType:    "text/plain",
			body:           `{"order": "1234567890"}`,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "missing content-type",
			contentType:    "",
			body:           `{"order": "1234567890"}`,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			h.RegisterOrder(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				require.Contains(t, w.Body.String(), "Content-Type must be application/json")
			}
		})
	}
}

func TestHandler_RegisterReward(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mock.NewMockOrderService(ctrl)
	mockReward := mock.NewMockRewardService(ctrl)
	h := handler.New(mockOrder, mockReward)

	tests := []struct {
		name           string
		contentType    string
		body           string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid JSON",
			contentType:    "application/json",
			body:           `{"sku": "ABC123", "reward_rate": 0.05}`,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid content-type",
			contentType:    "application/xml",
			body:           `{"sku": "ABC123"}`,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "missing content-type",
			contentType:    "",
			body:           `{"sku": "ABC123"}`,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/goods", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			h.RegisterReward(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				require.Contains(t, w.Body.String(), "Content-Type must be application/json")
			}
		})
	}
}
