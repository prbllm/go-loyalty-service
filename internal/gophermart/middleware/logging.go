package middleware

import (
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/prbllm/go-loyalty-service/internal/logger"
)

func Logging(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			log.Infof("%s %s %d %s", r.Method, r.URL.Path, ww.Status(), time.Since(start))
		})
	}
}
