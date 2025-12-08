package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
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
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}

	token, err := h.service.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		h.logger.Errorf("register error: %v", err)
		return
	}

	w.Header().Set(config.HeaderAuthorization, config.BearerPrefix+token)
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		h.logger.Errorf("login error: %v", err)
		return
	}

	w.Header().Set(config.HeaderAuthorization, config.BearerPrefix+token)
	w.WriteHeader(http.StatusOK)
}
