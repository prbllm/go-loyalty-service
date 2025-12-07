package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

func main() {
	appLogger, err := logger.NewZapLogger()
	if err != nil {
		fmt.Println("Error creating logger: ", err)
		os.Exit(1)
	}
	defer appLogger.Sync()

	err = config.InitConfig(config.GophermartFlagsSet)
	if err != nil {
		fmt.Println("Error initializing config: ", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	postgresRepository, err := repository.NewPostgresRepository(ctx, config.GetConfig().DatabaseURI, appLogger)
	if err != nil {
		fmt.Println("Error creating repository: ", err)
		os.Exit(1)
	}

	select {
	case <-ctx.Done():
		appLogger.Info("Received shutdown signal, shutting down server...")
	}

	_, shutdownCancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer shutdownCancel()

	if pgRepo, ok := postgresRepository.(*repository.PostgresRepository); ok {
		if closeErr := pgRepo.Close(); closeErr != nil {
			appLogger.Errorf("Error closing PostgreSQL connection: %v", closeErr)
		} else {
			appLogger.Info("PostgreSQL connection closed")
		}
	}
}
