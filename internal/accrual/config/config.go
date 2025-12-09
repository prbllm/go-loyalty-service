package config

import (
	"errors"
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress  string `env:"RUN_ADDRESS"`
	DatabaseURI string `env:"DATABASE_URI"`
}

func New(args []string) (*Config, error) {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	var (
		runAddress  = fs.String("a", "", "Accrual HTTP server address (e.g. localhost:8081)")
		databaseURI = fs.String("d", "", "Accrual database URI (e.g. postgres://accrual:accrual@localhost:5432/accrual?sslmode=disable)")
	)

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		RunAddress:  *runAddress,
		DatabaseURI: *databaseURI,
	}

	err = env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.RunAddress == "" {
		return nil, errors.New("required flag -a or env RUN_ADDRESS is missing")
	}
	if cfg.DatabaseURI == "" {
		return nil, errors.New("required flag -d or env DATABASE_URI is missing")
	}

	return cfg, nil
}
