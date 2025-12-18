package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/prbllm/go-loyalty-service/internal/accrual/handler"
	"github.com/prbllm/go-loyalty-service/internal/accrual/repository"
	"github.com/prbllm/go-loyalty-service/internal/accrual/service"
	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

func main() {
	appLogger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer appLogger.Sync()

	err = config.InitConfig(config.AccrualFlagsSet)
	if err != nil {
		appLogger.Fatal(err)
	}

	// Подключаемся к БД
	db, err := sql.Open("pgx", config.GetConfig().DatabaseURI)
	if err != nil {
		appLogger.Fatal(err)
	}
	defer db.Close()

	// Применяем миграции
	driver, err := postgres.WithInstance(db, &postgres.Config{MigrationsTable: "schema_migrations_accrual"})
	if err != nil {
		appLogger.Fatal(err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		"file://./migrations/accrual",
		"postgres", driver)
	if err != nil {
		appLogger.Fatal(err)
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		appLogger.Fatal(err)
	}

	// Создаём репозитории(уже на актуальной схеме!)
	orderRepo := repository.NewPostgresOrderRepo(db)
	rewardRepo := repository.NewPostgresRewardRepo(db)

	// Инициализируем сервисы
	orderService := service.NewOrderService(orderRepo, rewardRepo)
	rewardService := service.NewRewardService(rewardRepo)

	// Инициализируем обработчик
	h := handler.New(orderService, rewardService, appLogger)

	r := chi.NewRouter()
	r.Get("/api/orders/{number}", h.GetOrderInfo)
	r.Post("/api/orders", h.RegisterOrder)
	r.Post("/api/goods", h.RegisterReward)

	appLogger.Fatal(http.ListenAndServe(config.GetConfig().RunAddress, r))
}
