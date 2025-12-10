package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	JWTSecret            string
}

var globalConfig *Config

func defaultConfig() *Config {
	return &Config{
		RunAddress:           DefaultRunAddress,
		DatabaseURI:          DefaultDatabaseURI,
		AccrualSystemAddress: DefaultAccrualSystemAddress,
		JWTSecret:            DefaultJWTSecret,
	}
}

func InitConfig(flagsetName string) error {
	globalConfig = ParseFlags(flagsetName, os.Args[1:], flag.ExitOnError)
	globalConfig.loadFromEnvironment(flagsetName)
	return globalConfig.Validate(flagsetName)
}

func GetConfig() *Config {
	if globalConfig == nil {
		globalConfig = defaultConfig()
	}
	return globalConfig
}

func SetConfig(config *Config) {
	globalConfig = config
}

func (c *Config) Validate(flagsetName string) error {
	if c.RunAddress == "" {
		return fmt.Errorf("run address cannot be empty")
	}

	if c.DatabaseURI == "" {
		return fmt.Errorf("database URI cannot be empty")
	}

	if flagsetName == GophermartFlagsSet {
		if c.AccrualSystemAddress == "" {
			return fmt.Errorf("accrual system address cannot be empty")
		}
	}

	return nil
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{RunAddress: %s, DatabaseURI: %s, AccrualSystemAddress: %s, JWTSecret: %s}",
		c.RunAddress, c.DatabaseURI, c.AccrualSystemAddress, c.JWTSecret)
}

func (c *Config) loadFromEnvironment(flagsetName string) {
	if address, err := GetEnvironment(RunAddressEnv); err == nil {
		c.RunAddress = address
	}

	if databaseURI, err := GetEnvironment(DatabaseURIEnv); err == nil {
		c.DatabaseURI = databaseURI
	}

	if flagsetName == GophermartFlagsSet {
		if accrualSystemAddress, err := GetEnvironment(AccrualSystemAddressEnv); err == nil {
			c.AccrualSystemAddress = accrualSystemAddress
		}
	}
}
