package handler

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prbllm/go-loyalty-service/internal/accrual/model"
	"github.com/prbllm/go-loyalty-service/internal/accrual/service"
	"github.com/prbllm/go-loyalty-service/pkg/luhn"
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
	// Извлекаем номер заказа из URL
	number := chi.URLParam(r, "number")

	// Валидация: проходит алгоритм Луна
	if !luhn.IsValidOrderNumber(number) {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	// Запрашиваем заказ у сервиса
	order, err := h.orderService.GetOrder(r.Context(), number)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	orderResponse := model.GetOrderResponse{
		Number: order.Number,
		Status: string(order.Status),
	}

	if order.Accrual != nil {
		// Переводим из копеек в рубли
		accrualResutl := float64(*order.Accrual) / 100
		orderResponse.Accrual = &accrualResutl
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orderResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// POST /api/orders — регистрация нового совершённого заказа
func (h *Handler) RegisterOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var order model.RegisterOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация номера заказа (должен проходить алгоритм Луна)
	if !luhn.IsValidOrderNumber(order.Number) {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	// Валидация товаров
	if len(order.Goods) == 0 {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	var goods []model.Good
	for _, item := range order.Goods {
		if item.Description == "" || item.Price <= 0 {
			http.Error(w, "invalid request format", http.StatusBadRequest)
			return
		}

		// Переводим рубли в копейки: 47399.99 → 4739999
		// Округляем до ближайшего целого копейки (используем 2 знака)
		priceInCents := int64(math.Round(item.Price * 100))

		goods = append(goods, model.Good{
			Description: item.Description,
			Price:       priceInCents,
		})
	}

	newOrder := model.Order{
		Number: order.Number,
		Goods:  goods,
		Status: model.Registered,
	}

	err := h.orderService.RegisterOrder(r.Context(), newOrder)
	if err != nil {
		if errors.Is(err, service.ErrOrderAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if rewardRule.Reward <= 0 {
		http.Error(w, "invalid request format", http.StatusBadRequest)
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
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
