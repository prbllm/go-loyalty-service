package service

import (
	"errors"
	"testing"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/logger"
	mocks "github.com/prbllm/go-loyalty-service/internal/mocks/accrual"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_orderService_RegisterOrder(t *testing.T) {
	tests := []struct {
		name        string
		order       model.Order
		mockSetup   func(*mocks.MockOrderRepository, *mocks.MockRewardRepository)
		expectedErr error
	}{
		{
			name: "order exists error",
			order: model.Order{
				Number: "1234567890",
				Goods: []model.Good{
					{Description: "Чайник Bork", Price: 7000},
				},
			},
			mockSetup: func(m *mocks.MockOrderRepository, r *mocks.MockRewardRepository) {
				m.EXPECT().IsOrderExists(gomock.Any(), "1234567890").Return(false, errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
		{
			name: "order already exists",
			order: model.Order{
				Number: "1234567890",
				Goods: []model.Good{
					{Description: "Чайник Bork", Price: 7000},
				},
			},
			mockSetup: func(m *mocks.MockOrderRepository, r *mocks.MockRewardRepository) {
				m.EXPECT().IsOrderExists(gomock.Any(), "1234567890").Return(true, nil)
			},
			expectedErr: ErrOrderAlreadyExists,
		},
		{
			name: "order create error",
			order: model.Order{
				Number: "1234567890",
				Goods: []model.Good{
					{Description: "Чайник Bork", Price: 7000},
				},
				Status: model.Registered,
			},
			mockSetup: func(m *mocks.MockOrderRepository, r *mocks.MockRewardRepository) {
				m.EXPECT().IsOrderExists(gomock.Any(), "1234567890").Return(false, nil)
				m.EXPECT().Create(gomock.Any(), model.Order{Number: "1234567890", Goods: []model.Good{
					{Description: "Чайник Bork", Price: 7000},
				}, Status: model.Registered, Accrual: nil}).Return(errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrderRepo := mocks.NewMockOrderRepository(ctrl)
			mockRewardRepo := mocks.NewMockRewardRepository(ctrl)
			tt.mockSetup(mockOrderRepo, mockRewardRepo)

			logger := logger.NewNop()
			orderService := NewOrderService(mockOrderRepo, mockRewardRepo, logger)

			err := orderService.RegisterOrder(t.Context(), tt.order)
			require.Equal(t, err, tt.expectedErr)
		})
	}
}

func Test_orderService_GetOrder(t *testing.T) {
	tests := []struct {
		name        string
		number      string
		mockSetup   func(*mocks.MockOrderRepository)
		want        *model.Order
		expectedErr error
	}{
		{
			name:   "internal db error",
			number: "1234567890",
			mockSetup: func(m *mocks.MockOrderRepository) {
				m.EXPECT().IsOrderExists(gomock.Any(), "1234567890").Return(false, errors.New("db error"))
			},
			want:        nil,
			expectedErr: errors.New("db error"),
		},
		{
			name:   "order not found",
			number: "1234567890",
			mockSetup: func(m *mocks.MockOrderRepository) {
				m.EXPECT().IsOrderExists(gomock.Any(), "1234567890").Return(false, nil)
			},
			want:        nil,
			expectedErr: ErrOrderNotFound,
		},
		{
			name:   "getbynumber internal db error",
			number: "1234567890",
			mockSetup: func(m *mocks.MockOrderRepository) {
				m.EXPECT().IsOrderExists(gomock.Any(), "1234567890").Return(true, nil)
				m.EXPECT().GetByNumber(gomock.Any(), "1234567890").Return(nil, errors.New("db error"))
			},
			want:        nil,
			expectedErr: errors.New("db error"),
		},
		{
			name:   "order found",
			number: "1234567890",
			mockSetup: func(m *mocks.MockOrderRepository) {
				m.EXPECT().IsOrderExists(gomock.Any(), "1234567890").Return(true, nil)
				m.EXPECT().GetByNumber(gomock.Any(), "1234567890").Return(&model.Order{Number: "12345678900"}, nil)
			},
			want:        &model.Order{Number: "12345678900"},
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrderRepo := mocks.NewMockOrderRepository(ctrl)
			mockRewardRepo := mocks.NewMockRewardRepository(ctrl)
			tt.mockSetup(mockOrderRepo)

			logger := logger.NewNop()
			orderService := NewOrderService(mockOrderRepo, mockRewardRepo, logger)

			order, err := orderService.GetOrder(t.Context(), tt.number)
			require.Equal(t, order, tt.want)
			require.Equal(t, err, tt.expectedErr)
		})
	}
}
