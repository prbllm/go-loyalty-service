package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/middleware"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/service/balance"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/utils"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

type BalanceHandler struct {
	service balance.Service
	logger  logger.Logger
}

func NewBalanceHandler(service balance.Service, logger logger.Logger) *BalanceHandler {
	return &BalanceHandler{
		service: service,
		logger:  logger,
	}
}

type balanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type withdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h *BalanceHandler) Balance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	bal, err := h.service.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		h.logger.Errorf("balance: get balance: %v", err)
		return
	}

	resp := balanceResponse{
		Current:   bal.Current,
		Withdrawn: bal.Withdrawn,
	}

	w.Header().Set(config.HeaderContentType, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *BalanceHandler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	items, err := h.service.GetWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		h.logger.Errorf("withdrawals: get list: %v", err)
		return
	}

	if len(items) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]withdrawalResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, withdrawalResponse{
			Order:       it.OrderNumber,
			Sum:         it.Sum,
			ProcessedAt: it.ProcessedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set(config.HeaderContentType, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !utils.IsValidOrderNumber(req.Order) {
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		return
	}

	if req.Sum <= 0 {
		http.Error(w, "invalid sum", http.StatusBadRequest)
		return
	}

	err := h.service.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		switch err {
		case repository.ErrInsufficientFunds:
			http.Error(w, "insufficient funds", http.StatusPaymentRequired)
			return
		default:
			h.logger.Errorf("withdraw: user %d order %s: %v", userID, req.Order, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
