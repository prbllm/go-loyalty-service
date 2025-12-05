package config

import (
	"fmt"
	"os"
)

func GetEnvironment(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("environment variable %s is not set", key)
	}
	if value == "" {
		return "", fmt.Errorf("environment variable %s is set but empty", key)
	}
	return value, nil
}
