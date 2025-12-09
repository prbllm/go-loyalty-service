package config_test

import (
	"testing"

	"github.com/prbllm/go-loyalty-service/internal/accrual/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		envVars map[string]string
		want    *config.Config
		wantErr bool
	}{
		{
			name: "success with flags",
			args: []string{"-a", "localhost:8081", "-d", "postgres://user:pass@localhost/db"},
			want: &config.Config{
				RunAddress:  "localhost:8081",
				DatabaseURI: "postgres://user:pass@localhost/db",
			},
			wantErr: false,
		},
		{
			name:    "success with env vars",
			args:    []string{},
			envVars: map[string]string{"RUN_ADDRESS": "localhost:8081", "DATABASE_URI": "postgres://user:pass@localhost/db"},
			want: &config.Config{
				RunAddress:  "localhost:8081",
				DatabaseURI: "postgres://user:pass@localhost/db",
			},
			wantErr: false,
		},
		{
			name:    "env override flags vars",
			args:    []string{"-a", "localhost:8081", "-d", "postgres://user:pass@localhost/db"},
			envVars: map[string]string{"RUN_ADDRESS": "0.0.0.0:8082", "DATABASE_URI": "postgres://from-env"},
			want: &config.Config{
				RunAddress:  "0.0.0.0:8082",
				DatabaseURI: "postgres://from-env",
			},
			wantErr: false,
		},
		{
			name:    "missing RunAddress",
			args:    []string{"-d", "postgres://user:pass@localhost/db"},
			envVars: nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "missing DatabaseURI",
			args:    []string{"-a", "localhost:8081"},
			envVars: nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "missing both",
			args:    []string{},
			envVars: nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty flags",
			args:    []string{"-a", "", "-d", ""},
			envVars: nil,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем переменные окружения для текущего теста
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			got, err := config.New(tt.args)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
