package service

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/accrual/repository/mock"
	"github.com/stretchr/testify/require"
)

func Test_rewardService_RegisterReward(t *testing.T) {
	tests := []struct {
		name        string
		reward      model.RewardRule
		mockSetup   func(*mock.MockRewardRepository)
		expectedErr error
	}{
		{
			name: "ExistsByMatch error",
			reward: model.RewardRule{
				Match:      "Bork",
				Reward:     10,
				RewardType: model.RewardTypePercent,
			},
			mockSetup: func(m *mock.MockRewardRepository) {
				m.EXPECT().ExistsByMatch(gomock.Any(), "Bork").Return(false, errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
		{
			name: "ExistsByMatch",
			reward: model.RewardRule{
				Match:      "Bork",
				Reward:     10,
				RewardType: model.RewardTypePercent,
			},
			mockSetup: func(m *mock.MockRewardRepository) {
				m.EXPECT().ExistsByMatch(gomock.Any(), "Bork").Return(true, nil)
			},
			expectedErr: ErrMatchAlreadyExists,
		},
		{
			name: "not exists create error",
			reward: model.RewardRule{
				Match:      "Bork",
				Reward:     10,
				RewardType: model.RewardTypePercent,
			},
			mockSetup: func(m *mock.MockRewardRepository) {
				m.EXPECT().ExistsByMatch(gomock.Any(), "Bork").Return(false, nil)
				m.EXPECT().Create(gomock.Any(), gomock.Eq(model.RewardRule{Match: "Bork", Reward: 10, RewardType: model.RewardTypePercent})).Return(errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
		{
			name: "success create",
			reward: model.RewardRule{
				Match:      "Bork",
				Reward:     10,
				RewardType: model.RewardTypePercent,
			},
			mockSetup: func(m *mock.MockRewardRepository) {
				m.EXPECT().ExistsByMatch(gomock.Any(), "Bork").Return(false, nil)
				m.EXPECT().Create(gomock.Any(), gomock.Eq(model.RewardRule{Match: "Bork", Reward: 10, RewardType: model.RewardTypePercent})).Return(nil)
			},
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockRewardRepository(ctrl)
			tt.mockSetup(mockRepo)

			rewardService := NewRewardService(mockRepo)

			err := rewardService.RegisterReward(t.Context(), tt.reward)
			require.Equal(t, err, tt.expectedErr)
		})
	}
}
