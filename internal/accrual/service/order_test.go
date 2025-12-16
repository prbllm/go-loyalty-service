package service

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/accrual/repository/mock"
	"github.com/stretchr/testify/require"
)

func Test_orderService_RegisterOrder(t *testing.T) {
	tests := []struct {
		name        string
		order       model.Order
		mockSetup   func(*mock.MockOrderRepository)
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
			mockSetup: func(m *mock.MockOrderRepository) {
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
			mockSetup: func(m *mock.MockOrderRepository) {
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
			mockSetup: func(m *mock.MockOrderRepository) {
				m.EXPECT().IsOrderExists(gomock.Any(), "1234567890").Return(false, nil)
				m.EXPECT().Create(gomock.Any(), model.Order{Number: "1234567890", Goods: []model.Good{
					{Description: "Чайник Bork", Price: 7000},
				}, Status: model.Registered, Accrual: nil}).Return(errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
		{
			name: "order create successfully",
			order: model.Order{
				Number: "1234567890",
				Goods: []model.Good{
					{Description: "Чайник Bork", Price: 7000},
				},
				Status: model.Registered,
			},
			mockSetup: func(m *mock.MockOrderRepository) {
				m.EXPECT().IsOrderExists(gomock.Any(), "1234567890").Return(false, nil)
				m.EXPECT().Create(gomock.Any(), model.Order{Number: "1234567890", Goods: []model.Good{
					{Description: "Чайник Bork", Price: 7000},
				}, Status: model.Registered, Accrual: nil}).Return(nil)
			},
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrderRepo := mock.NewMockOrderRepository(ctrl)
			mockRewardRepo := mock.NewMockRewardRepository(ctrl)
			tt.mockSetup(mockOrderRepo)

			orderService := NewOrderService(mockOrderRepo, mockRewardRepo)

			err := orderService.RegisterOrder(t.Context(), tt.order)
			require.Equal(t, err, tt.expectedErr)
		})
	}
}
