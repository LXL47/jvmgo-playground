package api

import (
	"sync"
	"time"
)

type visitor struct {
	window time.Time
	count  int
}

type limiter struct {
	mu          sync.Mutex
	limit       int
	maxVisitors int
	lastCleanup time.Time
	visitors    map[string]visitor
}

func newLimiter(limit int) *limiter {
	return &limiter{limit: limit, maxVisitors: 10_000, visitors: make(map[string]visitor)}
}

func (l *limiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lastCleanup.IsZero() || now.Sub(l.lastCleanup) >= time.Minute {
		for address, current := range l.visitors {
			if now.Sub(current.window) >= 2*time.Minute {
				delete(l.visitors, address)
			}
		}
		l.lastCleanup = now
	}
	current := l.visitors[key]
	if current.window.IsZero() || now.Sub(current.window) >= time.Minute {
		if current.window.IsZero() && len(l.visitors) >= l.maxVisitors {
			return false
		}
		l.visitors[key] = visitor{window: now, count: 1}
		return true
	}
	if current.count >= l.limit {
		return false
	}
	current.count++
	l.visitors[key] = current
	return true
}
