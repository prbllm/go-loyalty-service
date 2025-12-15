package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/prbllm/go-loyalty-service/internal/accrual/config"
	"github.com/prbllm/go-loyalty-service/internal/accrual/handler"
	// "github.com/prbllm/go-loyalty-service/internal/accrual/repository"
)

func main() {
	cfg, err := config.New(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	// Подключаемся к БД
	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Применяем миграции
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		"file://./migrations/accrual",
		"postgres", driver)
	if err != nil {
		log.Fatal(err)
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	// Создаём репозитории(уже на актуальной схеме!)
	// orderRepo := repository.NewPostgresOrderRepo(db)
	// rewardRepo := repository.NewPostgresRewardRepo(db)

	// Продолжаем инициализацию сервиса
	h := handler.New()

	r := chi.NewRouter()
	r.Get("/api/orders/{number}", h.GetOrderInfo)
	r.Post("/api/orders", h.RegisterOrder)
	r.Post("/api/goods", h.RegisterReward)

	log.Fatal(http.ListenAndServe(cfg.RunAddress, r))
}
