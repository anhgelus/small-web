package backend

import (
	"context"
	"sync"
	"time"
)

type to struct {
	n     int
	since time.Time
}

type tos struct {
	mu  sync.Mutex
	tos map[string]*to
}

var timeouts = tos{tos: make(map[string]*to)}

func rateLimitDuration(n int) time.Duration {
	return time.Duration(1<<2*n) * time.Second
}

func isRateLimited(ctx context.Context) bool {
	ip := ContextIP(ctx)

	timeouts.mu.Lock()
	defer timeouts.mu.Unlock()

	v, ok := timeouts.tos[ip]
	if !ok {
		return false
	}
	return time.Since(v.since) <= rateLimitDuration(v.n)
}

func rateLimit(ctx context.Context) bool {
	ip := ContextIP(ctx)

	timeouts.mu.Lock()
	defer timeouts.mu.Unlock()

	v, ok := timeouts.tos[ip]
	if !ok {
		timeouts.tos[ip] = &to{n: 1}
		return false
	}
	if time.Since(v.since) <= rateLimitDuration(v.n) {
		return true
	}
	v.n++
	if v.n%4 != 0 {
		return false
	}
	v.since = time.Now()
	ContextLogger(ctx).Warn(
		"rate limiting IP",
		"ip", ip,
		"duration", rateLimitDuration(v.n).String())
	go func(v *to, ip string) {
		time.Sleep(3 * time.Hour)
		v.n = max(v.n-4, 0)
		if v.n == 0 {
			timeouts.mu.Lock()
			defer timeouts.mu.Unlock()
			delete(timeouts.tos, ip)
		}
	}(v, ip)
	return true
}

func resetRateLimit(ctx context.Context) {
	ip := ContextIP(ctx)

	timeouts.mu.Lock()
	defer timeouts.mu.Unlock()

	delete(timeouts.tos, ip)
}
