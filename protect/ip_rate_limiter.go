package protect

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

//
// ────────────────────────────────────────────────────────────
//  IP EXTRACTOR
// ────────────────────────────────────────────────────────────
//

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

//
// ────────────────────────────────────────────────────────────
//  RATE LIMITER INTERNAL STRUCTS
// ────────────────────────────────────────────────────────────
//

// Fixed-window limiter.
// entries always counts requests **only for the current window**.
// We reset it entirely when the window expires → leak-free by construction.
type IPRateLimiter struct {
	mu              sync.Mutex
	entries         map[string]int // count per IP
	maxRequests     int
	window          time.Duration
	currWindowStart time.Time
}

//
// ────────────────────────────────────────────────────────────
//  BUILDER
// ────────────────────────────────────────────────────────────
//

const (
	defaultMaxRequests = 10
	defaultWindow      = time.Second
)

type IPRateLimiterBuilder struct {
	maxRequests *int
	window      *time.Duration
}

func NewIPRateLimiter() *IPRateLimiterBuilder {
	return &IPRateLimiterBuilder{}
}

func (b *IPRateLimiterBuilder) WithDefaults() *IPRateLimiterBuilder {
	if b.maxRequests == nil {
		v := defaultMaxRequests
		b.maxRequests = &v
	}
	if b.window == nil {
		v := defaultWindow
		b.window = &v
	}
	return b
}

func (b *IPRateLimiterBuilder) MaxRequests(n int) *IPRateLimiterBuilder {
	b.maxRequests = &n
	return b
}

func (b *IPRateLimiterBuilder) Window(d time.Duration) *IPRateLimiterBuilder {
	b.window = &d
	return b
}

func (b *IPRateLimiterBuilder) Build() *IPRateLimiter {
	if b.maxRequests == nil {
		v := defaultMaxRequests
		b.maxRequests = &v
	}
	if b.window == nil {
		v := defaultWindow
		b.window = &v
	}
	if *b.maxRequests <= 0 {
		panic("protect: maxRequests must be > 0")
	}
	if *b.window <= 0 {
		panic("protect: window duration must be > 0")
	}

	return &IPRateLimiter{
		entries:         make(map[string]int),
		maxRequests:     *b.maxRequests,
		window:          *b.window,
		currWindowStart: time.Now(),
	}
}

//
// ────────────────────────────────────────────────────────────
//  RATE LIMITER LOGIC (fixed-window, leak-free)
// ────────────────────────────────────────────────────────────
//

// Allow implements a fixed-window rate limiter with **zero leak**.
// Every window reset discards all stored IP counts.
func (rl *IPRateLimiter) Allow(key string) bool {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Reset the entire window if expired
	if now.Sub(rl.currWindowStart) >= rl.window {
		rl.entries = make(map[string]int) // atomic purge → no leaks
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

//
// ────────────────────────────────────────────────────────────
//  INSPECT (debug only)
// ────────────────────────────────────────────────────────────
//

func (rl *IPRateLimiter) Inspect(key string) (count int, start time.Time, found bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	count, ok := rl.entries[key]
	return count, rl.currWindowStart, ok
}
