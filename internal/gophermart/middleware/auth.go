package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/prbllm/go-loyalty-service/internal/config"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/utils"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(config.HeaderAuthorization)
		if header == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], strings.TrimSpace(config.BearerPrefix)) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := utils.ParseToken(parts[1])
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	val := ctx.Value(userIDContextKey)
	userID, ok := val.(int64)
	return userID, ok
}
