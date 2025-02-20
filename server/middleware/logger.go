package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{ResponseWriter: w}
}

func (rw *responseWriterWrapper) Status() int {
	return rw.status
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Logger is a middleware that will log information about the HTTP request that was made.
func Logger(l *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := newResponseWriterWrapper(w)
			next.ServeHTTP(wrapped, r)

			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}

			duration := time.Since(start)

			l.Info(
				fmt.Sprintf("%s %s://%s%s from %s - %d in %s",
					r.Method,
					scheme,
					r.Host,
					r.URL.EscapedPath(),
					r.RemoteAddr,
					wrapped.status,
					duration,
				),
				"method", r.Method,
				"scheme", scheme,
				"host", r.Host,
				"path", r.URL.EscapedPath(),
				"remoteAddress", r.RemoteAddr,
				"status", wrapped.status,
				"duration", duration,
			)
		}
		return http.HandlerFunc(fn)
	}
}
