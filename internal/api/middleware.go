package api

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
)

type contextKey string

const keyCtx contextKey = "apiKey"

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get("Authorization")
		if raw == "" {
			Error(w, http.StatusUnauthorized, "missing Authorization header")
			return
		}

		raw = strings.TrimPrefix(raw, "Bearer ")
		raw = strings.TrimSpace(raw)

		key, err := s.apiKeySvc.ValidateKey(raw)
		if err != nil {
			Error(w, http.StatusUnauthorized, "invalid API key")
			return
		}

		ctx := context.WithValue(r.Context(), keyCtx, key)
		next(w, r.WithContext(ctx))
	}
}

func currentKey(r *http.Request) *domain.APIKey {
	key, _ := r.Context().Value(keyCtx).(*domain.APIKey)
	return key
}

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[uint]*bucket
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

var limiter = &rateLimiter{buckets: make(map[uint]*bucket)}

func (s *Server) trayAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Tray-Token")
		if token == "" || token != s.internalToken {
			Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		session, err := auth.LoadSession()
		if err != nil {
			Error(w, http.StatusUnauthorized, "no active session — please log in via sprintos start")
			return
		}

		user, err := s.userSvc.GetByEmail(session.Email)
		if err != nil {
			Error(w, http.StatusUnauthorized, "user not found")
			return
		}

		org, err := s.orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			Error(w, http.StatusUnauthorized, "organisation not found")
			return
		}

		syntheticKey := &domain.APIKey{
			UserID: user.ID,
			OrgID:  org.ID,
		}

		ctx := context.WithValue(r.Context(), keyCtx, syntheticKey)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) rateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := currentKey(r)
		if key == nil {
			next(w, r)
			return
		}

		limiter.mu.Lock()
		b, ok := limiter.buckets[key.ID]
		if !ok || time.Since(b.lastReset) > time.Minute {
			limiter.buckets[key.ID] = &bucket{tokens: 59, lastReset: time.Now()}
			limiter.mu.Unlock()
			next(w, r)
			return
		}

		if b.tokens <= 0 {
			limiter.mu.Unlock()
			w.Header().Set("Retry-After", "60")
			Error(w, http.StatusTooManyRequests, "rate limit exceeded — 60 requests per minute")
			return
		}

		b.tokens--
		limiter.mu.Unlock()
		next(w, r)
	}
}
