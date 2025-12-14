package accrual

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetOrder(t *testing.T) {
	tests := []struct {
		name            string
		statusCode      int
		body            string
		retryAfter      string
		assertErr       func(error) bool
		expectedStatus  string
		expectedAccrual float64
	}{
		{
			name:            "ok processed",
			statusCode:      http.StatusOK,
			body:            `{"order":"123","status":"PROCESSED","accrual":50}`,
			expectedStatus:  StatusProcessed,
			expectedAccrual: 50,
		},
		{
			name:       "no content",
			statusCode: http.StatusNoContent,
			assertErr: func(err error) bool {
				return errors.Is(err, ErrOrderNotRegistered)
			},
		},
		{
			name:       "too many requests",
			statusCode: http.StatusTooManyRequests,
			retryAfter: "3",
			assertErr: func(err error) bool {
				var tmr *TooManyRequestsError
				return errors.As(err, &tmr) && tmr.RetryAfter == 3*time.Second
			},
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			assertErr: func(err error) bool {
				return errors.Is(err, ErrUnexpectedStatus)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.retryAfter != "" {
					w.Header().Set("Retry-After", tt.retryAfter)
				}
				w.WriteHeader(tt.statusCode)
				if tt.body != "" {
					_, _ = w.Write([]byte(tt.body))
				}
			}))
			defer server.Close()

			client := NewClient(server.URL+"/", nil)

			resp, err := client.GetOrder(context.Background(), "123")

			if tt.assertErr != nil {
				if err == nil || !tt.assertErr(err) {
					t.Fatalf("expected error condition not met, got: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.Status != tt.expectedStatus {
				t.Fatalf("expected status %s, got %s", tt.expectedStatus, resp.Status)
			}

			if resp.Accrual != tt.expectedAccrual {
				t.Fatalf("expected accrual %v, got %v", tt.expectedAccrual, resp.Accrual)
			}
		})
	}
}
