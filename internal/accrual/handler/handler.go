package handler

import "net/http"

type Handler struct {
}

func New() *Handler {
	h := &Handler{}
	return h
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
	w.WriteHeader(http.StatusOK)
}
