package service

import (
	"errors"
	"testing"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	mocks "github.com/prbllm/go-loyalty-service/internal/mocks/accrual"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_rewardService_RegisterReward(t *testing.T) {
	tests := []struct {
		name        string
		reward      model.RewardRule
		mockSetup   func(*mocks.MockRewardRepository)
		expectedErr error
	}{
		{
			name: "ExistsByMatch error",
			reward: model.RewardRule{
				Match:      "Bork",
				Reward:     10,
				RewardType: model.RewardTypePercent,
			},
			mockSetup: func(m *mocks.MockRewardRepository) {
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
			mockSetup: func(m *mocks.MockRewardRepository) {
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
			mockSetup: func(m *mocks.MockRewardRepository) {
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
			mockSetup: func(m *mocks.MockRewardRepository) {
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

			mockRepo := mocks.NewMockRewardRepository(ctrl)
			tt.mockSetup(mockRepo)

			rewardService := NewRewardService(mockRepo)

			err := rewardService.RegisterReward(t.Context(), tt.reward)
			require.Equal(t, err, tt.expectedErr)
		})
	}
}
