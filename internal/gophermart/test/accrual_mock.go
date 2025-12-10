package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type AccrualMock struct {
	server     *http.Server
	mu         sync.RWMutex
	orders     map[string]AccrualResponse
	port       int
	tooMany    bool
	retryAfter time.Duration
}

func NewAccrualMock() (*AccrualMock, error) {
	port, err := findFreePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find free port: %w", err)
	}

	mock := &AccrualMock{
		orders:     make(map[string]AccrualResponse),
		port:       port,
		tooMany:    false,
		retryAfter: 1 * time.Second,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/orders/", mock.handleOrder)

	mock.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return mock, nil
}

func (m *AccrualMock) URL() string {
	return fmt.Sprintf("http://localhost:%d", m.port)
}

func (m *AccrualMock) SetOrder(orderNumber, status string, accrual float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders[orderNumber] = AccrualResponse{
		Order:   orderNumber,
		Status:  status,
		Accrual: accrual,
	}
}

func (m *AccrualMock) SetTooManyRequests(enabled bool, retryAfter time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tooMany = enabled
	m.retryAfter = retryAfter
}

func (m *AccrualMock) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders = make(map[string]AccrualResponse)
	m.tooMany = false
}

func (m *AccrualMock) Start() error {
	return m.server.ListenAndServe()
}

func (m *AccrualMock) Stop(ctx context.Context) error {
	return m.server.Shutdown(ctx)
}

func (m *AccrualMock) handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	m.mu.RLock()
	tooMany := m.tooMany
	retryAfter := m.retryAfter
	m.mu.RUnlock()

	if tooMany {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())))
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	orderNumber := r.URL.Path[len("/api/orders/"):]
	if orderNumber == "" {
		http.Error(w, "order number required", http.StatusBadRequest)
		return
	}

	m.mu.RLock()
	response, exists := m.orders[orderNumber]
	m.mu.RUnlock()

	if !exists {
		http.Error(w, "order not found", http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
