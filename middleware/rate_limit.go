package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	mu       sync.Mutex
	limiters map[int32]*rate.Limiter
	r        rate.Limit
	burst    int
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[int32]*rate.Limiter),
		r:        rate.Limit(rps),
		burst:    burst,
	}
}

func (rl *RateLimiter) getLimiter(operatorID int32) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[operatorID]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.burst)
		rl.limiters[operatorID] = limiter
	}

	return limiter
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		operator, ok := OperatorFromContext(r.Context())
		if !ok {
			http.Error(w, "missing operator", http.StatusUnauthorized)
			return
		}

		limiter := rl.getLimiter(operator.ID)

		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
