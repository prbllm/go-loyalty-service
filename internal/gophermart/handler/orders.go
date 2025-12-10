package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/middleware"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/service/order"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

type OrderHandler struct {
	service order.Service
	logger  logger.Logger
}

type orderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func NewOrderHandler(service order.Service, logger logger.Logger) *OrderHandler {
	return &OrderHandler{
		service: service,
		logger:  logger,
	}
}

func (h *OrderHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	number := strings.TrimSpace(string(body))
	if number == "" {
		http.Error(w, "invalid order number", http.StatusBadRequest)
		return
	}

	err = h.service.Upload(r.Context(), userID, number)
	if err != nil {
		switch {
		case err == order.ErrInvalidOrderNumber:
			http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
			return
		case err == order.ErrOrderAlreadyUploadedBySameUser:
			w.WriteHeader(http.StatusOK)
			return
		case err == order.ErrOrderUploadedByAnotherUser:
			http.Error(w, "order belongs to another user", http.StatusConflict)
			return
		default:
			h.logger.Errorf("upload order error: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.service.List(r.Context(), userID)
	if err != nil {
		h.logger.Errorf("list orders error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		var accrual *float64
		if o.Status == model.OrderStatusProcessed {
			accrual = &o.Accrual
		}

		response = append(response, orderResponse{
			Number:     o.Number,
			Status:     o.Status,
			Accrual:    accrual,
			UploadedAt: o.UploadedAt,
		})
	}

	w.Header().Set(config.HeaderContentType, config.ContentTypeJSON)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorf("encode orders response: %v", err)
	}
}
