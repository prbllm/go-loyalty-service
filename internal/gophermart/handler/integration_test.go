package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prbllm/go-loyalty-service/internal/config"
	gmiddleware "github.com/prbllm/go-loyalty-service/internal/gophermart/middleware"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/utils"
	mocks "github.com/prbllm/go-loyalty-service/internal/mocks/gophermart"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestIntegrationHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := zaptest.NewLogger(t).Sugar()

	authSvc := mocks.NewMockAuthService(ctrl)
	orderSvc := mocks.NewMockOrderService(ctrl)
	balanceSvc := mocks.NewMockBalanceService(ctrl)

	authSvc.EXPECT().Register(gomock.Any(), "user", "pass").Return("token-register", nil)
	authSvc.EXPECT().Login(gomock.Any(), "user", "pass").Return("token-login", nil)

	orderSvc.EXPECT().Upload(gomock.Any(), int64(1), "79927398713").Return(nil)
	now := time.Now().UTC()
	orderSvc.EXPECT().List(gomock.Any(), int64(1)).Return([]*model.Order{
		{
			Number:     "79927398713",
			Status:     model.OrderStatusProcessed,
			Accrual:    model.Amount(1050),
			UploadedAt: now,
		},
	}, nil)

	balanceSvc.EXPECT().GetBalance(gomock.Any(), int64(1)).Return(&model.Balance{
		Current:   model.Amount(1050),
		Withdrawn: model.Amount(0),
	}, nil)
	balanceSvc.EXPECT().Withdraw(gomock.Any(), int64(1), "79927398713", model.Amount(500)).Return(nil)
	balanceSvc.EXPECT().GetWithdrawals(gomock.Any(), int64(1)).Return([]*model.Withdrawal{
		{OrderNumber: "79927398713", Sum: model.Amount(500), ProcessedAt: now},
	}, nil)

	router := chi.NewRouter()
	router.Use(
		chimiddleware.Compress(5),
		gmiddleware.Logging(log),
	)

	authHandler := NewAuthHandler(authSvc, log)
	orderHandler := NewOrderHandler(orderSvc, log)
	balanceHandler := NewBalanceHandler(balanceSvc, log)

	router.Post(config.PathUserRegister, authHandler.Register)
	router.Post(config.PathUserLogin, authHandler.Login)
	router.With(gmiddleware.Auth).Post(config.PathUserOrders, orderHandler.Upload)
	router.With(gmiddleware.Auth).Get(config.PathUserOrders, orderHandler.List)
	router.With(gmiddleware.Auth).Get(config.PathUserBalance, balanceHandler.Balance)
	router.With(gmiddleware.Auth).Post(config.PathUserWithdraw, balanceHandler.Withdraw)
	router.With(gmiddleware.Auth).Get(config.PathWithdrawals, balanceHandler.Withdrawals)

	server := httptest.NewServer(router)
	defer server.Close()

	client := server.Client()

	resp := doJSONRequest(t, client, http.MethodPost, server.URL+config.PathUserRegister, `{"login":"user","password":"pass"}`, "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("register expected 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get(config.HeaderAuthorization); got != config.BearerPrefix+"token-register" {
		t.Fatalf("register expected auth header, got %s", got)
	}

	resp = doJSONRequest(t, client, http.MethodPost, server.URL+config.PathUserLogin, `{"login":"user","password":"pass"}`, "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login expected 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get(config.HeaderAuthorization); got != config.BearerPrefix+"token-login" {
		t.Fatalf("login expected auth header, got %s", got)
	}

	token, _ := utils.GenerateToken(1)
	authHeader := config.BearerPrefix + token

	resp = doRequest(t, client, http.MethodPost, server.URL+config.PathUserOrders, "79927398713", authHeader)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("upload expected 202, got %d", resp.StatusCode)
	}

	req, err := http.NewRequest(http.MethodGet, server.URL+config.PathUserOrders, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set(config.HeaderAuthorization, authHeader)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("orders expected 200, got %d", resp.StatusCode)
	}
	if ce := resp.Header.Get("Content-Encoding"); ce != "gzip" {
		t.Fatalf("expected gzip encoding, got %s", ce)
	}
	body := ungzipBody(t, resp.Body)
	var orders []map[string]interface{}
	if err := json.Unmarshal(body, &orders); err != nil {
		t.Fatalf("unmarshal orders: %v", err)
	}
	if len(orders) != 1 || orders[0]["number"] != "79927398713" {
		t.Fatalf("unexpected orders response: %v", orders)
	}

	resp = doRequest(t, client, http.MethodGet, server.URL+config.PathUserBalance, "", authHeader)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("balance expected 200, got %d", resp.StatusCode)
	}

	resp = doJSONRequest(t, client, http.MethodPost, server.URL+config.PathUserWithdraw, `{"order":"79927398713","sum":5}`, authHeader)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("withdraw expected 200, got %d", resp.StatusCode)
	}

	resp = doRequest(t, client, http.MethodGet, server.URL+config.PathWithdrawals, "", authHeader)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("withdrawals expected 200, got %d", resp.StatusCode)
	}
}

func doJSONRequest(t *testing.T, client *http.Client, method, url, body, authHeader string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if authHeader != "" {
		req.Header.Set(config.HeaderAuthorization, authHeader)
	}
	req.Header.Set(config.HeaderContentType, config.ContentTypeJSON)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func doRequest(t *testing.T, client *http.Client, method, url, body, authHeader string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if authHeader != "" {
		req.Header.Set(config.HeaderAuthorization, authHeader)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func ungzipBody(t *testing.T, r io.ReadCloser) []byte {
	t.Helper()
	defer r.Close()
	gr, err := gzip.NewReader(r)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer gr.Close()
	data, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("read gzip body: %v", err)
	}
	return data
}
