package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     time.Duration
	burst    int
}

// Visitor tracks rate limit state for an IP
type Visitor struct {
	tokens   int
	lastSeen time.Time
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: how often to add tokens (e.g., 100ms for 10 req/sec)
// burst: maximum tokens (burst capacity)
func NewRateLimiter(rate time.Duration, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		burst:    burst,
	}

	// Clean up old visitors every 5 minutes
	go rl.cleanupVisitors()

	return rl
}

// getVisitor returns or creates a visitor for an IP
func (rl *RateLimiter) getVisitor(ip string) *Visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &Visitor{
			tokens:   rl.burst,
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
	}

	return v
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	visitor := rl.getVisitor(ip)

	visitor.mu.Lock()
	defer visitor.mu.Unlock()

	// Add tokens based on time passed
	now := time.Now()
	elapsed := now.Sub(visitor.lastSeen)
	tokensToAdd := int(elapsed / rl.rate)

	if tokensToAdd > 0 {
		visitor.tokens += tokensToAdd
		if visitor.tokens > rl.burst {
			visitor.tokens = rl.burst
		}
		visitor.lastSeen = now
	}

	// Check if we have tokens available
	if visitor.tokens > 0 {
		visitor.tokens--
		return true
	}

	return false
}

// cleanupVisitors removes visitors that haven't been seen in 5 minutes
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			v.mu.Lock()
			if time.Since(v.lastSeen) > 5*time.Minute {
				delete(rl.visitors, ip)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware returns a middleware that enforces rate limiting
func RateLimitMiddleware(rl *RateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get IP address
			ip := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}

			if !rl.Allow(ip) {
				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next(w, r)
		}
	}
}

// APIKeyAuth implements simple API key authentication for admin operations
type APIKeyAuth struct {
	apiKey string
}

// NewAPIKeyAuth creates a new API key authenticator
func NewAPIKeyAuth(apiKey string) *APIKeyAuth {
	return &APIKeyAuth{apiKey: apiKey}
}

// AuthMiddleware returns a middleware that checks API key for protected endpoints
func (a *APIKeyAuth) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for GET requests (read-only)
		if r.Method == http.MethodGet {
			next(w, r)
			return
		}

		// Check API key for write operations
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			apiKey = r.URL.Query().Get("api_key")
		}

		// If no API key configured, allow (backward compatibility)
		if a.apiKey == "" {
			next(w, r)
			return
		}

		if apiKey != a.apiKey {
			http.Error(w, "Unauthorized: Invalid or missing API key", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// LoggingMiddleware logs all requests
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call next handler
		next(w, r)

		// Log request
		duration := time.Since(start)
		// In production, use proper logging library
		_ = duration // Suppress unused warning
		// log.Printf("%s %s - %v", r.Method, r.URL.Path, duration)
	}
}

// CompressionMiddleware adds gzip compression for responses
func CompressionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next(w, r)
			return
		}

		// Create gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Wrap response writer
		gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		gzw.Header().Set("Content-Encoding", "gzip")
		gzw.Header().Del("Content-Length") // Let gzip set this

		next(gzw, r)
	}
}

// gzipResponseWriter wraps http.ResponseWriter with gzip compression
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
