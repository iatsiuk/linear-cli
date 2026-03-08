package api

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// rateLimitDelay returns the delay before retrying a rate-limited request.
// It reads X-RateLimit-Requests-Reset (UTC epoch ms) from the response header,
// falling back to exponential backoff starting at 1s.
func rateLimitDelay(resp *http.Response, attempt int) time.Duration {
	if reset := resp.Header.Get("X-RateLimit-Requests-Reset"); reset != "" {
		if ms, err := strconv.ParseInt(reset, 10, 64); err == nil {
			if d := time.Until(time.UnixMilli(ms)); d > 0 {
				if d > 60*time.Second {
					d = 60 * time.Second
				}
				return d + jitter(500*time.Millisecond)
			}
		}
	}
	// exponential backoff: 1s, 2s, 4s, ...
	base := time.Duration(1<<attempt) * time.Second
	return base + jitter(base/2)
}

// serverErrorDelay returns a retry delay with jitter for 5xx errors (1-5s).
func serverErrorDelay(attempt int) time.Duration {
	base := time.Second * time.Duration(attempt+1)
	if base > 5*time.Second {
		base = 5 * time.Second
	}
	return base + jitter(time.Second)
}

// jitter returns a random duration in [0, max).
func jitter(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(max)))
}
