package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/service/auth"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/service/order"
)

type errorResponse struct {
	Error string `json:"error"`
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch {
	case errors.Is(err, repository.ErrUserAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, auth.ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, order.ErrInvalidOrderNumber):
		return http.StatusUnprocessableEntity
	case errors.Is(err, order.ErrOrderAlreadyUploadedBySameUser):
		return http.StatusOK
	case errors.Is(err, order.ErrOrderUploadedByAnotherUser):
		return http.StatusConflict
	case errors.Is(err, repository.ErrInsufficientFunds):
		return http.StatusPaymentRequired
	default:
		return http.StatusInternalServerError
	}
}

func getErrorMessage(err error, statusCode int) string {
	if err == nil {
		return ""
	}

	switch statusCode {
	case http.StatusConflict:
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return "user already exists"
		}
		if errors.Is(err, order.ErrOrderUploadedByAnotherUser) {
			return "order belongs to another user"
		}
		return "conflict"
	case http.StatusUnauthorized:
		return "invalid credentials"
	case http.StatusUnprocessableEntity:
		return "invalid order number"
	case http.StatusPaymentRequired:
		return "insufficient funds"
	default:
		return "internal error"
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set(config.HeaderContentType, config.ContentTypeJSON)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse{Error: message})
}
