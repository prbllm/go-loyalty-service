package test

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestRegister_Success(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	reqBody := `{"login":"testuser","password":"testpass"}`
	req, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserRegister, strings.NewReader(reqBody))
	require.NoError(t, err)
	req.Header.Set(config.HeaderContentType, config.ContentTypeJSON)

	resp, err := env.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	token := resp.Header.Get(config.HeaderAuthorization)
	assert.NotEmpty(t, token)
	assert.True(t, strings.HasPrefix(token, config.BearerPrefix))
}

func TestRegister_Duplicate(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	reqBody := `{"login":"duplicate","password":"pass"}`
	req, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserRegister, strings.NewReader(reqBody))
	require.NoError(t, err)
	req.Header.Set(config.HeaderContentType, config.ContentTypeJSON)

	resp, err := env.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req2, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserRegister, strings.NewReader(reqBody))
	require.NoError(t, err)
	req2.Header.Set(config.HeaderContentType, config.ContentTypeJSON)

	resp2, err := env.httpClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusConflict, resp2.StatusCode)
}

func TestRegister_InvalidFormat(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	tests := []struct {
		name     string
		body     string
		wantCode int
	}{
		{"empty body", "", http.StatusBadRequest},
		{"invalid json", `{`, http.StatusBadRequest},
		{"empty login", `{"login":"","password":"pass"}`, http.StatusBadRequest},
		{"empty password", `{"login":"user","password":""}`, http.StatusBadRequest},
		{"missing fields", `{"login":"user"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserRegister, strings.NewReader(tt.body))
			require.NoError(t, err)
			if tt.body != "" {
				req.Header.Set(config.HeaderContentType, config.ContentTypeJSON)
			}

			resp, err := env.httpClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantCode, resp.StatusCode)
		})
	}
}

func TestLogin_Success(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	registerBody := `{"login":"loginuser","password":"loginpass"}`
	req, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserRegister, strings.NewReader(registerBody))
	require.NoError(t, err)
	req.Header.Set(config.HeaderContentType, config.ContentTypeJSON)

	resp, err := env.httpClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	loginReq, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserLogin, strings.NewReader(registerBody))
	require.NoError(t, err)
	loginReq.Header.Set(config.HeaderContentType, config.ContentTypeJSON)

	loginResp, err := env.httpClient.Do(loginReq)
	require.NoError(t, err)
	defer loginResp.Body.Close()

	assert.Equal(t, http.StatusOK, loginResp.StatusCode)
	token := loginResp.Header.Get(config.HeaderAuthorization)
	assert.NotEmpty(t, token)
	assert.True(t, strings.HasPrefix(token, config.BearerPrefix))
}

func TestLogin_InvalidCredentials(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	reqBody := `{"login":"nonexistent","password":"wrong"}`
	req, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserLogin, strings.NewReader(reqBody))
	require.NoError(t, err)
	req.Header.Set(config.HeaderContentType, config.ContentTypeJSON)

	resp, err := env.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func makeAuthRequest(t *testing.T, env *testEnvironment, method, path, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, env.gophermartURL+path, strings.NewReader(body))
	require.NoError(t, err)
	if body != "" && strings.HasPrefix(body, "{") {
		req.Header.Set(config.HeaderContentType, config.ContentTypeJSON)
	}

	token := getAuthToken(t, env)
	if token != "" {
		req.Header.Set(config.HeaderAuthorization, config.BearerPrefix+token)
	}

	resp, err := env.httpClient.Do(req)
	require.NoError(t, err)
	return resp
}

func getAuthToken(t *testing.T, env *testEnvironment) string {
	t.Helper()
	reqBody := `{"login":"authtest","password":"authtest"}`

	loginReq, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserLogin, strings.NewReader(reqBody))
	require.NoError(t, err)
	loginReq.Header.Set(config.HeaderContentType, config.ContentTypeJSON)

	loginResp, err := env.httpClient.Do(loginReq)
	require.NoError(t, err)
	defer loginResp.Body.Close()

	if loginResp.StatusCode == http.StatusOK {
		token := loginResp.Header.Get(config.HeaderAuthorization)
		return strings.TrimPrefix(token, config.BearerPrefix)
	}

	req, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserRegister, strings.NewReader(reqBody))
	require.NoError(t, err)
	req.Header.Set(config.HeaderContentType, config.ContentTypeJSON)

	resp, err := env.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		token := resp.Header.Get(config.HeaderAuthorization)
		return strings.TrimPrefix(token, config.BearerPrefix)
	}

	return ""
}

func TestOrderUpload_Success(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	env.accrualMock.SetOrder("79927398713", "REGISTERED", 0)

	resp := makeAuthRequest(t, env, http.MethodPost, config.PathUserOrders, "79927398713")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestOrderUpload_DuplicateSameUser(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	orderNumber := "79927398713"
	env.accrualMock.SetOrder(orderNumber, "REGISTERED", 0)

	resp1 := makeAuthRequest(t, env, http.MethodPost, config.PathUserOrders, orderNumber)
	resp1.Body.Close()
	assert.Equal(t, http.StatusAccepted, resp1.StatusCode)

	resp2 := makeAuthRequest(t, env, http.MethodPost, config.PathUserOrders, orderNumber)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

func TestOrderUpload_DuplicateDifferentUser(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	orderNumber := "79927398713"
	env.accrualMock.SetOrder(orderNumber, "REGISTERED", 0)

	user1Body := `{"login":"user1","password":"pass1"}`
	req1, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserRegister, strings.NewReader(user1Body))
	require.NoError(t, err)
	req1.Header.Set(config.HeaderContentType, config.ContentTypeJSON)

	resp1, err := env.httpClient.Do(req1)
	require.NoError(t, err)
	resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	token1 := strings.TrimPrefix(resp1.Header.Get(config.HeaderAuthorization), config.BearerPrefix)

	orderReq1, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserOrders, strings.NewReader(orderNumber))
	require.NoError(t, err)
	orderReq1.Header.Set(config.HeaderAuthorization, config.BearerPrefix+token1)

	orderResp1, err := env.httpClient.Do(orderReq1)
	require.NoError(t, err)
	orderResp1.Body.Close()
	assert.Equal(t, http.StatusAccepted, orderResp1.StatusCode)

	user2Body := `{"login":"user2","password":"pass2"}`
	req2, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserRegister, strings.NewReader(user2Body))
	require.NoError(t, err)
	req2.Header.Set(config.HeaderContentType, config.ContentTypeJSON)

	resp2, err := env.httpClient.Do(req2)
	require.NoError(t, err)
	resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
	token2 := strings.TrimPrefix(resp2.Header.Get(config.HeaderAuthorization), config.BearerPrefix)

	orderReq2, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserOrders, strings.NewReader(orderNumber))
	require.NoError(t, err)
	orderReq2.Header.Set(config.HeaderAuthorization, config.BearerPrefix+token2)

	orderResp2, err := env.httpClient.Do(orderReq2)
	require.NoError(t, err)
	defer orderResp2.Body.Close()
	assert.Equal(t, http.StatusConflict, orderResp2.StatusCode)
}

func TestOrderUpload_InvalidNumber(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	resp := makeAuthRequest(t, env, http.MethodPost, config.PathUserOrders, "123")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestOrderUpload_Unauthorized(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	req, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserOrders, strings.NewReader("79927398713"))
	require.NoError(t, err)

	resp, err := env.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestOrderList_Empty(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	resp := makeAuthRequest(t, env, http.MethodGet, config.PathUserOrders, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestOrderList_WithOrders(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	orderNumber := "79927398713"
	env.accrualMock.SetOrder(orderNumber, "PROCESSED", 100.5)

	resp := makeAuthRequest(t, env, http.MethodPost, config.PathUserOrders, orderNumber)
	resp.Body.Close()

	time.Sleep(2 * time.Second)

	listResp := makeAuthRequest(t, env, http.MethodGet, config.PathUserOrders, "")
	defer listResp.Body.Close()

	assert.Equal(t, http.StatusOK, listResp.StatusCode)
	assert.Equal(t, config.ContentTypeJSON, listResp.Header.Get(config.HeaderContentType))

	var orders []map[string]interface{}
	err := json.NewDecoder(listResp.Body).Decode(&orders)
	require.NoError(t, err)
	assert.Greater(t, len(orders), 0)
}

func TestOrderList_Gzip(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	orderNumber := "79927398713"
	env.accrualMock.SetOrder(orderNumber, "PROCESSED", 100.5)

	resp := makeAuthRequest(t, env, http.MethodPost, config.PathUserOrders, orderNumber)
	resp.Body.Close()

	time.Sleep(2 * time.Second)

	req, err := http.NewRequest(http.MethodGet, env.gophermartURL+config.PathUserOrders, nil)
	require.NoError(t, err)
	token := getAuthToken(t, env)
	req.Header.Set(config.HeaderAuthorization, config.BearerPrefix+token)
	req.Header.Set("Accept-Encoding", "gzip")

	resp2, err := env.httpClient.Do(req)
	require.NoError(t, err)
	defer resp2.Body.Close()

	if resp2.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(resp2.Body)
		require.NoError(t, err)
		defer gr.Close()
		data, err := io.ReadAll(gr)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
	}
}

func TestBalance_Success(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	resp := makeAuthRequest(t, env, http.MethodGet, config.PathUserBalance, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, config.ContentTypeJSON, resp.Header.Get(config.HeaderContentType))

	var balance map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&balance)
	require.NoError(t, err)
	assert.Contains(t, balance, "current")
	assert.Contains(t, balance, "withdrawn")
}

func TestBalance_Unauthorized(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	req, err := http.NewRequest(http.MethodGet, env.gophermartURL+config.PathUserBalance, nil)
	require.NoError(t, err)

	resp, err := env.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWithdraw_Success(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	orderNumber := "79927398713"
	env.accrualMock.SetOrder(orderNumber, "PROCESSED", 1000.0)

	resp := makeAuthRequest(t, env, http.MethodPost, config.PathUserOrders, orderNumber)
	resp.Body.Close()

	time.Sleep(2 * time.Second)

	withdrawBody := fmt.Sprintf(`{"order":"2377225624","sum":100}`)
	withdrawResp := makeAuthRequest(t, env, http.MethodPost, config.PathUserWithdraw, withdrawBody)
	defer withdrawResp.Body.Close()

	assert.Equal(t, http.StatusOK, withdrawResp.StatusCode)
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	withdrawBody := `{"order":"2377225624","sum":1000}`
	resp := makeAuthRequest(t, env, http.MethodPost, config.PathUserWithdraw, withdrawBody)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusPaymentRequired, resp.StatusCode)
}

func TestWithdraw_InvalidOrder(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	withdrawBody := `{"order":"123","sum":100}`
	resp := makeAuthRequest(t, env, http.MethodPost, config.PathUserWithdraw, withdrawBody)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestWithdrawals_Empty(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	resp := makeAuthRequest(t, env, http.MethodGet, config.PathWithdrawals, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestWithdrawals_WithData(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	orderNumber := "79927398713"
	env.accrualMock.SetOrder(orderNumber, "PROCESSED", 1000.0)

	resp := makeAuthRequest(t, env, http.MethodPost, config.PathUserOrders, orderNumber)
	resp.Body.Close()

	time.Sleep(2 * time.Second)

	withdrawBody := fmt.Sprintf(`{"order":"2377225624","sum":100}`)
	withdrawResp := makeAuthRequest(t, env, http.MethodPost, config.PathUserWithdraw, withdrawBody)
	withdrawResp.Body.Close()

	withdrawalsResp := makeAuthRequest(t, env, http.MethodGet, config.PathWithdrawals, "")
	defer withdrawalsResp.Body.Close()

	assert.Equal(t, http.StatusOK, withdrawalsResp.StatusCode)
	assert.Equal(t, config.ContentTypeJSON, withdrawalsResp.Header.Get(config.HeaderContentType))

	var withdrawals []map[string]interface{}
	err := json.NewDecoder(withdrawalsResp.Body).Decode(&withdrawals)
	require.NoError(t, err)
	assert.Greater(t, len(withdrawals), 0)
}

func TestFullFlow(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.teardown()

	registerBody := `{"login":"flowuser","password":"flowpass"}`
	registerReq, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserRegister, strings.NewReader(registerBody))
	require.NoError(t, err)
	registerReq.Header.Set(config.HeaderContentType, config.ContentTypeJSON)

	registerResp, err := env.httpClient.Do(registerReq)
	require.NoError(t, err)
	registerResp.Body.Close()
	assert.Equal(t, http.StatusOK, registerResp.StatusCode)

	token := registerResp.Header.Get(config.HeaderAuthorization)
	token = strings.TrimPrefix(token, config.BearerPrefix)

	orderNumber := "79927398713"
	env.accrualMock.SetOrder(orderNumber, "PROCESSED", 500.0)

	orderReq, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserOrders, strings.NewReader(orderNumber))
	require.NoError(t, err)
	orderReq.Header.Set(config.HeaderAuthorization, config.BearerPrefix+token)

	orderResp, err := env.httpClient.Do(orderReq)
	require.NoError(t, err)
	orderResp.Body.Close()
	assert.Equal(t, http.StatusAccepted, orderResp.StatusCode)

	time.Sleep(2 * time.Second)

	balanceReq, err := http.NewRequest(http.MethodGet, env.gophermartURL+config.PathUserBalance, nil)
	require.NoError(t, err)
	balanceReq.Header.Set(config.HeaderAuthorization, config.BearerPrefix+token)

	balanceResp, err := env.httpClient.Do(balanceReq)
	require.NoError(t, err)
	defer balanceResp.Body.Close()
	assert.Equal(t, http.StatusOK, balanceResp.StatusCode)

	var balance map[string]interface{}
	err = json.NewDecoder(balanceResp.Body).Decode(&balance)
	require.NoError(t, err)
	current := balance["current"].(float64)
	assert.Greater(t, current, 0.0)

	withdrawBody := fmt.Sprintf(`{"order":"2377225624","sum":100}`)
	withdrawReq, err := http.NewRequest(http.MethodPost, env.gophermartURL+config.PathUserWithdraw, strings.NewReader(withdrawBody))
	require.NoError(t, err)
	withdrawReq.Header.Set(config.HeaderAuthorization, config.BearerPrefix+token)
	withdrawReq.Header.Set(config.HeaderContentType, config.ContentTypeJSON)

	withdrawResp, err := env.httpClient.Do(withdrawReq)
	require.NoError(t, err)
	withdrawResp.Body.Close()
	assert.Equal(t, http.StatusOK, withdrawResp.StatusCode)

	withdrawalsReq, err := http.NewRequest(http.MethodGet, env.gophermartURL+config.PathWithdrawals, nil)
	require.NoError(t, err)
	withdrawalsReq.Header.Set(config.HeaderAuthorization, config.BearerPrefix+token)

	withdrawalsResp, err := env.httpClient.Do(withdrawalsReq)
	require.NoError(t, err)
	defer withdrawalsResp.Body.Close()
	assert.Equal(t, http.StatusOK, withdrawalsResp.StatusCode)
}
