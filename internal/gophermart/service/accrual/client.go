package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prbllm/go-loyalty-service/internal/config"
)

const (
	StatusRegistered = "REGISTERED"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
	StatusInvalid    = "INVALID"

	defaultHTTPTimeout = 5 * time.Second
)

var (
	ErrOrderNotRegistered = errors.New("order not registered in accrual system")
	ErrUnexpectedStatus   = errors.New("unexpected status code from accrual system")
)

type TooManyRequestsError struct {
	RetryAfter time.Duration
}

func (e *TooManyRequestsError) Error() string {
	return fmt.Sprintf("too many requests, retry after %s", e.RetryAfter)
}

type Client interface {
	GetOrder(ctx context.Context, number string) (*Response, error)
}

type client struct {
	baseURL    string
	httpClient *http.Client
}

type Response struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func NewClient(baseURL string, httpClient *http.Client) Client {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	} else if httpClient.Timeout == 0 {
		httpClient.Timeout = defaultHTTPTimeout
	}

	return &client{
		baseURL:    trimmed,
		httpClient: httpClient,
	}
}

func (c *client) GetOrder(ctx context.Context, number string) (*Response, error) {
	url := fmt.Sprintf("%s%s/%s", c.baseURL, config.AccrualOrdersPath, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result Response
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &result, nil
	case http.StatusNoContent:
		return nil, ErrOrderNotRegistered
	case http.StatusTooManyRequests:
		retryAfter := parseRetryAfter(resp.Header.Get(config.HeaderRetryAfter))
		return nil, &TooManyRequestsError{RetryAfter: retryAfter}
	default:
		if resp.StatusCode >= http.StatusInternalServerError {
			return nil, fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
		}
		return nil, fmt.Errorf("accrual returned status %d", resp.StatusCode)
	}
}

func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}

	seconds, err := strconv.Atoi(header)
	if err != nil || seconds < 0 {
		return 0
	}

	return time.Duration(seconds) * time.Second
}
