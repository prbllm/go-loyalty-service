package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/accrual/service"
)

type Handler struct {
	orderService  service.OrderService
	rewardService service.RewardService
}

func New(orderService service.OrderService, rewardService service.RewardService) *Handler {
	return &Handler{orderService: orderService, rewardService: rewardService}
}

// GET /api/orders/{number} — получение информации о расчёте начислений баллов лояльности
func (h *Handler) GetOrderInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// POST /api/orders — регистрация нового совершённого заказа
func (h *Handler) RegisterOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// POST /api/goods — регистрация информации о новой механике вознаграждения за товар
func (h *Handler) RegisterReward(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var rewardRule model.RewardRule
	if err := json.NewDecoder(r.Body).Decode(&rewardRule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if rewardRule.Match == "" {
		http.Error(w, "'invalid request format", http.StatusBadRequest)
		return
	}

	if rewardRule.RewardType != model.RewardTypePercent && rewardRule.RewardType != model.RewardTypePoints {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	err := h.rewardService.RegisterReward(r.Context(), rewardRule)
	if err != nil {
		if errors.Is(err, service.ErrMatchAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	w.WriteHeader(http.StatusOK)
}
