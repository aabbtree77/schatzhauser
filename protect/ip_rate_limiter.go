package protect

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Fixed-window limiter. We reset entries entirely once the window expires.

const (
	defaultMaxRequests = 3
	defaultWindow      = 60 * time.Second
	defaultEnable      = false
)

type IPRateLimiter struct {
	mu              sync.Mutex
	entries         map[string]int // count per IP
	maxRequests     int
	window          time.Duration
	currWindowStart time.Time
	Enable          bool //this one exported to enable/disable the limiter inside handlers
}

type IPRateLimiterConfig struct {
	Enable      bool
	MaxRequests int
	Window      time.Duration
}

func NewIPRateLimiter(cfg IPRateLimiterConfig) *IPRateLimiter {
	rl := &IPRateLimiter{}

	// defaults
	/*
		Enable we just set directly as we cannot discern whether cfg.Enable
		was not set at all or specified as false.
		So if inside *.toml enable is true and we miss to pass it into cfg
		before applying this function, it will be set to false, sadly.
		Inside tomlSect2IPRConfig in routes.go I commented on this as well.
	*/

	rl.Enable = cfg.Enable

	if cfg.MaxRequests > 0 {
		rl.maxRequests = cfg.MaxRequests
	} else {
		rl.maxRequests = defaultMaxRequests
	}

	if cfg.Window > 0 {
		rl.window = cfg.Window
	} else {
		rl.window = defaultWindow
	}

	rl.entries = make(map[string]int)
	rl.currWindowStart = time.Now()

	return rl
}

// Allow implements a fixed-window rate limiter with **zero memory leaks**.
// Every window reset discards all stored IP counts.
func (rl *IPRateLimiter) Allow(key string) bool {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Reset the entire window if expired
	if now.Sub(rl.currWindowStart) >= rl.window {
		rl.entries = make(map[string]int) // atomic purge -> no leaks
		rl.currWindowStart = now
	}

	// Get current count
	count := rl.entries[key]

	if count >= rl.maxRequests {
		return false
	}

	// Increment
	rl.entries[key] = count + 1
	return true
}

// For debugging
func (rl *IPRateLimiter) Inspect(key string) (count int, start time.Time, found bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	count, ok := rl.entries[key]
	return count, rl.currWindowStart, ok
}

func GetIP(r *http.Request) string {
	hostPort := r.RemoteAddr
	if hostPort == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return strings.Trim(hostPort, "[]")
	}
	return host
}
