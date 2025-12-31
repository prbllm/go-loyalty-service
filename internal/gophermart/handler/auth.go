package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/service/auth"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

type AuthHandler struct {
	service auth.Service
	logger  logger.Logger
}

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func NewAuthHandler(service auth.Service, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "login and password are required")
		return
	}

	token, err := h.service.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		statusCode := getStatusCode(err)
		if statusCode == http.StatusInternalServerError {
			h.logger.Errorf("register error: %v", err)
		}
		writeJSONError(w, statusCode, getErrorMessage(err, statusCode))
		return
	}

	w.Header().Set(config.HeaderAuthorization, config.BearerPrefix+token)
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "login and password are required")
		return
	}

	token, err := h.service.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		statusCode := getStatusCode(err)
		if statusCode == http.StatusInternalServerError {
			h.logger.Errorf("login error: %v", err)
		}
		writeJSONError(w, statusCode, getErrorMessage(err, statusCode))
		return
	}

	w.Header().Set(config.HeaderAuthorization, config.BearerPrefix+token)
	w.WriteHeader(http.StatusOK)
}
