package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type clientBucket struct {
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
}

func LeakyBucketMiddleware(cfg Config, zapLog *zap.Logger) func(http.Handler) http.Handler {
	if !cfg.ThrottleEnabled {
		return func(next http.Handler) http.Handler { return next }
	}

	if zapLog == nil {
		zapLog = zap.NewNop()
	}

	ratePerSecond := cfg.ThrottleRatePerSec
	burstTokens := float64(cfg.ThrottleBurst)

	if ratePerSecond <= 0 || burstTokens <= 0 {
		zapLog.Warn("throttle disabled due to invalid config values",
			zap.Float64("rate_per_sec", ratePerSecond),
			zap.Int64("burst", cfg.ThrottleBurst),
		)
		return func(next http.Handler) http.Handler { return next }
	}

	cleanupInterval := cfg.ThrottleCleanupInterval
	expiration := cfg.ThrottleExpiration

	buckets := make(map[string]*clientBucket)
	var mutex sync.Mutex
	lastCleanupTime := time.Now()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			now := time.Now()
			clientKey := extractClientKey(request)

			mutex.Lock()
			defer mutex.Unlock()

			if cleanupInterval > 0 && now.Sub(lastCleanupTime) >= cleanupInterval && expiration > 0 {
				for key, bucket := range buckets {
					if now.Sub(bucket.lastSeen) > expiration {
						delete(buckets, key)
					}
				}
				lastCleanupTime = now
			}

			bucket, exists := buckets[clientKey]
			if !exists {
				bucket = &clientBucket{
					tokens:     burstTokens,
					lastRefill: now,
					lastSeen:   now,
				}
				buckets[clientKey] = bucket
			}

			elapsedSeconds := now.Sub(bucket.lastRefill).Seconds()
			if elapsedSeconds > 0 {
				bucket.tokens += elapsedSeconds * ratePerSecond
				if bucket.tokens > burstTokens {
					bucket.tokens = burstTokens
				}
				bucket.lastRefill = now
			}

			bucket.lastSeen = now

			if bucket.tokens < 1 {
				responseWriter.WriteHeader(http.StatusTooManyRequests)
				return
			}

			bucket.tokens -= 1
			next.ServeHTTP(responseWriter, request)
		})
	}
}

func extractClientKey(request *http.Request) string {
	forwardedFor := strings.TrimSpace(request.Header.Get("X-Forwarded-For"))
	if forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			candidate := strings.TrimSpace(parts[0])
			if candidate != "" {
				return candidate
			}
		}
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(request.RemoteAddr))
	if err == nil && host != "" {
		return host
	}

	if request.RemoteAddr != "" {
		return request.RemoteAddr
	}

	return "unknown"
}
