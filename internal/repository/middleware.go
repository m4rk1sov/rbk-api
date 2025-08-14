package repository

import (
	"github.com/m4rk1sov/rbk-api/pkg/jsonlog"
	"net/http"
	"time"
)

func RequestLogger(logger *jsonlog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.PrintInfo("request handled", map[string]string{
				"method": r.Method,
				"path":   r.URL.Path,
				"time":   time.Since(start).String(),
			})
		})
	}
}
