package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger logs HTTP requests and their statuses.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		
		// For A09 Security Logging: Log 401s, 403s, and 500s distinctly
		if rw.status >= 400 {
			log.Printf("[SECURITY] method=%s path=%s ip=%s status=%d duration=%v", r.Method, r.URL.Path, r.RemoteAddr, rw.status, duration)
		} else {
			log.Printf("method=%s path=%s ip=%s status=%d duration=%v", r.Method, r.URL.Path, r.RemoteAddr, rw.status, duration)
		}
	})
}
