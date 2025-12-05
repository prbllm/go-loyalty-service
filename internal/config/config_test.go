package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected *Config
	}{
		{
			name:   "default values",
			config: defaultConfig(),
			expected: &Config{
				RunAddress:           DefaultRunAddress,
				DatabaseURI:          DefaultDatabaseURI,
				AccrualSystemAddress: DefaultAccrualSystemAddress,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected.RunAddress, tt.config.RunAddress)
			assert.Equal(t, tt.expected.DatabaseURI, tt.config.DatabaseURI)
			assert.Equal(t, tt.expected.AccrualSystemAddress, tt.config.AccrualSystemAddress)
		})
	}
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name         string
		flagsetName  string
		args         []string
		expected     *Config
		expectedAddr string
		expectedDB   string
		expectedAcc  string
	}{
		{
			name:         "parse accrual flags",
			flagsetName:  AccrualFlagsSet,
			args:         []string{"-a", ":9090", "-d", "postgres://localhost/test"},
			expectedAddr: ":9090",
			expectedDB:   "postgres://localhost/test",
			expectedAcc:  "",
		},
		{
			name:         "parse gophermart flags",
			flagsetName:  GophermartFlagsSet,
			args:         []string{"-a", ":8080", "-d", "postgres://localhost/gophermart", "-r", "http://localhost:8081"},
			expectedAddr: ":8080",
			expectedDB:   "postgres://localhost/gophermart",
			expectedAcc:  "http://localhost:8081",
		},
		{
			name:         "parse flags with defaults",
			flagsetName:  AccrualFlagsSet,
			args:         []string{},
			expectedAddr: DefaultRunAddress,
			expectedDB:   DefaultDatabaseURI,
			expectedAcc:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseFlags(tt.flagsetName, tt.args, flag.ContinueOnError)
			assert.Equal(t, tt.expectedAddr, config.RunAddress)
			assert.Equal(t, tt.expectedDB, config.DatabaseURI)
			assert.Equal(t, tt.expectedAcc, config.AccrualSystemAddress)
		})
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	tests := []struct {
		name         string
		flagsetName  string
		envVars      map[string]string
		expectedAddr string
		expectedDB   string
		expectedAcc  string
	}{
		{
			name:        "load accrual environment",
			flagsetName: AccrualFlagsSet,
			envVars: map[string]string{
				RunAddressEnv:  ":9090",
				DatabaseURIEnv: "postgres://localhost/accrual",
			},
			expectedAddr: ":9090",
			expectedDB:   "postgres://localhost/accrual",
			expectedAcc:  "",
		},
		{
			name:        "load gophermart environment",
			flagsetName: GophermartFlagsSet,
			envVars: map[string]string{
				RunAddressEnv:           ":8080",
				DatabaseURIEnv:          "postgres://localhost/gophermart",
				AccrualSystemAddressEnv: "http://localhost:8081",
			},
			expectedAddr: ":8080",
			expectedDB:   "postgres://localhost/gophermart",
			expectedAcc:  "http://localhost:8081",
		},
		{
			name:        "environment overrides flags",
			flagsetName: AccrualFlagsSet,
			envVars: map[string]string{
				RunAddressEnv:  ":9999",
				DatabaseURIEnv: "postgres://override/db",
			},
			expectedAddr: ":9999",
			expectedDB:   "postgres://override/db",
			expectedAcc:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			config := defaultConfig()
			config.loadFromEnvironment(tt.flagsetName)

			assert.Equal(t, tt.expectedAddr, config.RunAddress)
			assert.Equal(t, tt.expectedDB, config.DatabaseURI)
			assert.Equal(t, tt.expectedAcc, config.AccrualSystemAddress)
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		flagsetName string
		wantErr     bool
		errMsg      string
	}{
		{
			name: "valid accrual config",
			config: &Config{
				RunAddress:  ":8080",
				DatabaseURI: "postgres://localhost/test",
			},
			flagsetName: AccrualFlagsSet,
			wantErr:     false,
		},
		{
			name: "valid gophermart config",
			config: &Config{
				RunAddress:           ":8080",
				DatabaseURI:          "postgres://localhost/test",
				AccrualSystemAddress: "http://localhost:8081",
			},
			flagsetName: GophermartFlagsSet,
			wantErr:     false,
		},
		{
			name: "empty run address",
			config: &Config{
				RunAddress:  "",
				DatabaseURI: "postgres://localhost/test",
			},
			flagsetName: AccrualFlagsSet,
			wantErr:     true,
			errMsg:      "run address cannot be empty",
		},
		{
			name: "empty database URI",
			config: &Config{
				RunAddress:  ":8080",
				DatabaseURI: "",
			},
			flagsetName: AccrualFlagsSet,
			wantErr:     true,
			errMsg:      "database URI cannot be empty",
		},
		{
			name: "empty accrual system address for gophermart",
			config: &Config{
				RunAddress:           ":8080",
				DatabaseURI:          "postgres://localhost/test",
				AccrualSystemAddress: "",
			},
			flagsetName: GophermartFlagsSet,
			wantErr:     true,
			errMsg:      "accrual system address cannot be empty",
		},
		{
			name: "empty accrual system address for accrual is ok",
			config: &Config{
				RunAddress:           ":8080",
				DatabaseURI:          "postgres://localhost/test",
				AccrualSystemAddress: "",
			},
			flagsetName: AccrualFlagsSet,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate(tt.flagsetName)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	t.Run("returns default config when global is nil", func(t *testing.T) {
		globalConfig = nil
		config := GetConfig()
		assert.NotNil(t, config)
		assert.Equal(t, DefaultRunAddress, config.RunAddress)
		assert.Equal(t, DefaultDatabaseURI, config.DatabaseURI)
	})

	t.Run("returns global config when set", func(t *testing.T) {
		expected := &Config{
			RunAddress:  ":9999",
			DatabaseURI: "postgres://test",
		}
		SetConfig(expected)
		config := GetConfig()
		assert.Equal(t, expected, config)
	})
}

func TestSetConfig(t *testing.T) {
	t.Run("sets global config", func(t *testing.T) {
		expected := &Config{
			RunAddress:  ":7777",
			DatabaseURI: "postgres://set",
		}
		SetConfig(expected)
		assert.Equal(t, expected, globalConfig)
	})
}
