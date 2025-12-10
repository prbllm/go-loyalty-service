package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/handler"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/middleware"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/service/accrual"
	authservice "github.com/prbllm/go-loyalty-service/internal/gophermart/service/auth"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/service/balance"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/service/order"
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

	repo, err := repository.NewPostgresRepository(ctx, config.GetConfig().DatabaseURI, appLogger)
	if err != nil {
		fmt.Println("Error creating repository: ", err)
		os.Exit(1)
	}

	accrualClient := accrual.NewClient(config.GetConfig().AccrualSystemAddress, nil)
	poller := accrual.NewWorkerPool(repo, accrualClient, appLogger, 0, 5)
	go poller.Run(ctx)

	authSvc := authservice.New(repo, appLogger)
	authHandler := handler.NewAuthHandler(authSvc, appLogger)
	orderSvc := order.New(repo, appLogger)
	orderHandler := handler.NewOrderHandler(orderSvc, appLogger)
	balanceSvc := balance.New(repo, appLogger)
	balanceHandler := handler.NewBalanceHandler(balanceSvc, appLogger)

	router := chi.NewRouter()
	router.Use(
		chimiddleware.Compress(5),
		middleware.Logging(appLogger),
	)
	router.Post(config.PathUserRegister, authHandler.Register)
	router.Post(config.PathUserLogin, authHandler.Login)
	router.With(middleware.Auth).Post(config.PathUserOrders, orderHandler.Upload)
	router.With(middleware.Auth).Get(config.PathUserOrders, orderHandler.List)
	router.With(middleware.Auth).Get(config.PathUserBalance, balanceHandler.Balance)
	router.With(middleware.Auth).Post(config.PathUserWithdraw, balanceHandler.Withdraw)
	router.With(middleware.Auth).Get(config.PathWithdrawals, balanceHandler.Withdrawals)

	srv := &http.Server{
		Addr:    config.GetConfig().RunAddress,
		Handler: router,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	appLogger.Infof("Server started on %s", config.GetConfig().RunAddress)

	select {
	case <-ctx.Done():
		appLogger.Info("Received shutdown signal, shutting down server...")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatalf("server error: %v", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Errorf("Server shutdown error: %v", err)
	}

	if closeErr := repo.Close(); closeErr != nil {
		appLogger.Errorf("Error closing repository: %v", closeErr)
	} else {
		appLogger.Info("Repository closed")
	}
}
