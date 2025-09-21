package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ResponseWriter wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.written {
		rw.statusCode = 200 // Default to 200 if WriteHeader wasn't called
		rw.written = true
	}
	return rw.ResponseWriter.Write(data)
}

// LoggingMiddleware logs all HTTP requests with method, path, status code, and duration
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // Default status
			written:        false,
		}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Determine log emoji based on status code
		var emoji string
		switch {
		case wrapped.statusCode >= 200 && wrapped.statusCode < 300:
			emoji = "✅"
		case wrapped.statusCode >= 300 && wrapped.statusCode < 400:
			emoji = "🔄"
		case wrapped.statusCode >= 400 && wrapped.statusCode < 500:
			emoji = "❌"
		case wrapped.statusCode >= 500:
			emoji = "💥"
		default:
			emoji = "❓"
		}

		// Enhanced logging for API routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			fmt.Printf("%s API %s %s -> %d (%v)\n",
				emoji, r.Method, r.URL.Path, wrapped.statusCode, duration)

			// Log query parameters if present
			if r.URL.RawQuery != "" {
				fmt.Printf("   📋 Query: %s\n", r.URL.RawQuery)
			}
		} else {
			// Regular logging for non-API routes
			fmt.Printf("%s %s %s -> %d (%v)\n",
				emoji, r.Method, r.URL.Path, wrapped.statusCode, duration)
		}
	})
}
