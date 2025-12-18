package handler_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prbllm/go-loyalty-service/internal/accrual/handler"
	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/accrual/service"
	mocks "github.com/prbllm/go-loyalty-service/internal/mocks/accrual"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHandler_GetOrderInfo(t *testing.T) {
	tests := []struct {
		name           string
		orderNumber    string
		expectedStatus int
		expectedBody   string
		mockSetup      func(*mocks.MockOrderService)
	}{
		{
			name:           "empty order number",
			orderNumber:    "",
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(m *mocks.MockOrderService) {},
		},
		{
			name:           "order not found",
			orderNumber:    "5354354162584",
			expectedStatus: http.StatusNoContent,
			mockSetup: func(m *mocks.MockOrderService) {
				m.EXPECT().GetOrder(gomock.Any(), "5354354162584").Return(nil, service.ErrOrderNotFound)
			},
		},
		{
			name:           "get order internal error",
			orderNumber:    "5354354162584",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func(m *mocks.MockOrderService) {
				m.EXPECT().GetOrder(gomock.Any(), "5354354162584").Return(nil, errors.New("service error"))
			},
		},
		{
			name:           "order with accrual",
			orderNumber:    "5354354162584",
			expectedStatus: http.StatusOK,
			mockSetup: func(m *mocks.MockOrderService) {
				accrual := int64(500)
				m.EXPECT().GetOrder(gomock.Any(), "5354354162584").Return(&model.Order{Number: "5354354162584", Status: model.Processed, Accrual: &accrual}, nil)
			},
			expectedBody: `{"order":"5354354162584","status":"PROCESSED","accrual":5}`,
		},
		{
			name:           "order without accrual",
			orderNumber:    "5354354162584",
			expectedStatus: http.StatusOK,
			mockSetup: func(m *mocks.MockOrderService) {
				m.EXPECT().GetOrder(gomock.Any(), "5354354162584").Return(&model.Order{Number: "5354354162584", Status: model.Processing}, nil)
			},
			expectedBody: `{"order":"5354354162584","status":"PROCESSING"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrder := mocks.NewMockOrderService(ctrl)
			mockReward := mocks.NewMockRewardService(ctrl)
			tt.mockSetup(mockOrder)

			h := handler.New(mockOrder, mockReward)

			req := httptest.NewRequest(http.MethodGet, "/api/orders/"+tt.orderNumber, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("number", tt.orderNumber)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			h.GetOrderInfo(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				require.JSONEq(t, w.Body.String(), tt.expectedBody)
				require.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestHandler_RegisterOrder(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		body           string
		expectedStatus int
		mockSetup      func(*mocks.MockOrderService)
	}{
		{
			name:           "invalid content-type",
			contentType:    "application/xml",
			body:           `{"order": "1234567890"}`,
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(m *mocks.MockOrderService) {},
		},
		{
			name:           "invalid json",
			contentType:    "application/json",
			body:           `{"order": "1234567890", "goods": [ {"description": "Чайник Bork", "price": 7000} ]`,
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(m *mocks.MockOrderService) {},
		},
		{
			name:           "invalid order",
			contentType:    "application/json",
			body:           `{"order": "", "goods": [ {"description": "Чайник Bork", "price": 7000} ]}`,
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(m *mocks.MockOrderService) {},
		},
		{
			name:           "invalid goods",
			contentType:    "application/json",
			body:           `{"order": "5354354162584", "goods": []}`,
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(m *mocks.MockOrderService) {},
		},
		{
			name:           "invalid description",
			contentType:    "application/json",
			body:           `{"order": "5354354162584", "goods": [ {"description": "", "price": 7000}]}`,
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(m *mocks.MockOrderService) {},
		},
		{
			name:           "invalid price",
			contentType:    "application/json",
			body:           `{"order": "5354354162584", "goods": [ {"description": "Чайник Bork", "price": 0}]}`,
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(m *mocks.MockOrderService) {},
		},
		{
			name:           "order already exists",
			contentType:    "application/json",
			body:           `{"order": "5354354162584", "goods": [ {"description": "Чайник Bork", "price": 7000}]}`,
			expectedStatus: http.StatusConflict,
			mockSetup: func(m *mocks.MockOrderService) {
				expectedOrder := model.Order{
					Number: "5354354162584",
					Goods: []model.Good{
						{Description: "Чайник Bork", Price: 700000},
					},
					Status: model.Registered,
				}
				m.EXPECT().RegisterOrder(gomock.Any(), gomock.Eq(expectedOrder)).Return(service.ErrOrderAlreadyExists)
			},
		},
		{
			name:           "internal service error",
			contentType:    "application/json",
			body:           `{"order": "5354354162584", "goods": [ {"description": "Чайник Bork", "price": 7000}]}`,
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func(m *mocks.MockOrderService) {
				expectedOrder := model.Order{
					Number: "5354354162584",
					Goods: []model.Good{
						{Description: "Чайник Bork", Price: 700000},
					},
					Status: model.Registered,
				}
				m.EXPECT().RegisterOrder(gomock.Any(), gomock.Eq(expectedOrder)).Return(errors.New("service error"))
			},
		},
		{
			name:           "order has been successfully accepted for processing",
			contentType:    "application/json",
			body:           `{"order": "5354354162584", "goods": [ {"description": "Чайник Bork", "price": 7000}]}`,
			expectedStatus: http.StatusAccepted,
			mockSetup: func(m *mocks.MockOrderService) {
				expectedOrder := model.Order{
					Number: "5354354162584",
					Goods: []model.Good{
						{Description: "Чайник Bork", Price: 700000},
					},
					Status: model.Registered,
				}
				m.EXPECT().RegisterOrder(gomock.Any(), gomock.Eq(expectedOrder)).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrder := mocks.NewMockOrderService(ctrl)
			mockReward := mocks.NewMockRewardService(ctrl)
			tt.mockSetup(mockOrder)

			h := handler.New(mockOrder, mockReward)

			req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			h.RegisterOrder(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_RegisterReward(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		body           string
		expectedStatus int
		mockSetup      func(*mocks.MockRewardService)
	}{
		{
			name:           "invalid content-type",
			contentType:    "application/xml",
			body:           `{"match": "Bork", "reward": 10, "reward_type": "%"}`,
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(m *mocks.MockRewardService) {},
		},
		{
			name:           "invalid json",
			contentType:    "application/json",
			body:           `{"match": "Bork", "reward": 10, "reward_type": "%"`,
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(m *mocks.MockRewardService) {},
		},
		{
			name:           "empty match",
			contentType:    "application/json",
			body:           `{"match": "", "reward": 10, "reward_type": "%"}`,
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(m *mocks.MockRewardService) {},
		},
		{
			name:           "invalid reward_type",
			contentType:    "application/json",
			body:           `{"match": "Bork", "reward": 10, "reward_type": "$"}`,
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(m *mocks.MockRewardService) {},
		},
		{
			name:           "invalid reward",
			contentType:    "application/json",
			body:           `{"match": "Bork", "reward": -10, "reward_type": "$"}`,
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func(m *mocks.MockRewardService) {},
		},
		{
			name:           "match alreay exists",
			contentType:    "application/json",
			body:           `{"match": "Bork", "reward": 10, "reward_type": "%"}`,
			expectedStatus: http.StatusConflict,
			mockSetup: func(m *mocks.MockRewardService) {
				expectedRewardRule := model.RewardRule{Match: "Bork", Reward: 10, RewardType: model.RewardTypePercent}
				m.EXPECT().RegisterReward(gomock.Any(), gomock.Eq(expectedRewardRule)).Return(service.ErrMatchAlreadyExists)
			},
		},
		{
			name:           "internal server error",
			contentType:    "application/json",
			body:           `{"match": "Bork", "reward": 10, "reward_type": "%"}`,
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func(m *mocks.MockRewardService) {
				expectedRewardRule := model.RewardRule{Match: "Bork", Reward: 10, RewardType: model.RewardTypePercent}
				m.EXPECT().RegisterReward(gomock.Any(), gomock.Eq(expectedRewardRule)).Return(errors.New("db error"))
			},
		},
		{
			name:           "valid JSON",
			contentType:    "application/json",
			body:           `{"match": "Bork", "reward": 10, "reward_type": "%"}`,
			expectedStatus: http.StatusOK,
			mockSetup: func(m *mocks.MockRewardService) {
				expectedRewardRule := model.RewardRule{Match: "Bork", Reward: 10, RewardType: model.RewardTypePercent}
				m.EXPECT().RegisterReward(gomock.Any(), gomock.Eq(expectedRewardRule)).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrder := mocks.NewMockOrderService(ctrl)
			mockReward := mocks.NewMockRewardService(ctrl)
			tt.mockSetup(mockReward)

			h := handler.New(mockOrder, mockReward)

			req := httptest.NewRequest(http.MethodPost, "/api/goods", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			h.RegisterReward(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
