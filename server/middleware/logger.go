package middleware

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"time"
)

// Logger is a middleware that will log information about the HTTP request that was made.
func Logger(l *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(wrapped, r)

			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}

			duration := time.Since(start)

			l.InfoContext(
				r.Context(),
				fmt.Sprintf("%s %s://%s%s from %s - %d in %s",
					r.Method,
					scheme,
					r.Host,
					r.URL.EscapedPath(),
					r.RemoteAddr,
					wrapped.Status(),
					duration,
				),
				"method", r.Method,
				"scheme", scheme,
				"host", r.Host,
				"path", r.URL.EscapedPath(),
				"remoteAddress", r.RemoteAddr,
				"status", wrapped.Status(),
				"duration", duration,
			)
		}
		return http.HandlerFunc(fn)
	}
}
