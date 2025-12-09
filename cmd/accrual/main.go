package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/prbllm/go-loyalty-service/internal/accrual/config"
	"github.com/prbllm/go-loyalty-service/internal/accrual/handler"
)

func main() {
	cfg, err := config.New(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	h := handler.New()

	r := chi.NewRouter()
	r.Get("/api/orders/{number}", h.GetOrderInfo)
	r.Post("/api/orders", h.RegisterOrder)
	r.Post("/api/goods", h.RegisterReward)

	log.Fatal(http.ListenAndServe(cfg.RunAddress, r))
}
