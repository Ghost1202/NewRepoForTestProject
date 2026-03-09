package interceptor

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type tokenBucket struct {
	tokensCount  float64
	lastRefillAt time.Time
}

func RateLimit(config Config, logger *zap.Logger) grpc.UnaryServerInterceptor {
	if !config.RateLimitEnabled || config.RateLimitRPS <= 0 || config.RateLimitBurst <= 0 {
		return func(serverContext context.Context, serverRequest any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(serverContext, serverRequest)
		}
	}

	type rateLimiterStorage struct {
		sync.Mutex
		buckets map[string]*tokenBucket
	}

	storage := &rateLimiterStorage{
		buckets: make(map[string]*tokenBucket),
	}

	limitRatePerSecond := float64(config.RateLimitRPS)
	limitBurstSize := float64(config.RateLimitBurst)

	go func() {
		ticker := time.NewTicker(config.RateLimitCleanup)
		defer ticker.Stop()

		for range ticker.C {
			storage.Lock()
			currentTime := time.Now()
			for clientIP, userBucket := range storage.buckets {
				if currentTime.Sub(userBucket.lastRefillAt) > config.RateLimitTTL {
					delete(storage.buckets, clientIP)
				}
			}
			storage.Unlock()
		}
	}()

	return func(serverContext context.Context, serverRequest any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		clientIP := "unknown_client"

		if peerInfo, ok := peer.FromContext(serverContext); ok {
			host, _, err := net.SplitHostPort(peerInfo.Addr.String())
			if err == nil {
				clientIP = host
			} else {
				clientIP = peerInfo.Addr.String()
			}
		}

		if incomingMetadata, ok := metadata.FromIncomingContext(serverContext); ok {
			if headerValues := incomingMetadata.Get("x-forwarded-for"); len(headerValues) > 0 {

				clientIP = strings.Split(headerValues[0], ",")[0]
				clientIP = strings.TrimSpace(clientIP)
			} else if headerValues := incomingMetadata.Get("x-real-ip"); len(headerValues) > 0 {
				clientIP = strings.TrimSpace(headerValues[0])
			}
		}

		storage.Lock()

		userBucket, exists := storage.buckets[clientIP]
		currentTime := time.Now()

		if !exists {
			userBucket = &tokenBucket{
				tokensCount:  limitBurstSize,
				lastRefillAt: currentTime,
			}
			storage.buckets[clientIP] = userBucket
		}

		elapsedTime := currentTime.Sub(userBucket.lastRefillAt).Seconds()
		newTokens := elapsedTime * limitRatePerSecond

		userBucket.tokensCount += newTokens
		if userBucket.tokensCount > limitBurstSize {
			userBucket.tokensCount = limitBurstSize
		}
		userBucket.lastRefillAt = currentTime

		isAllowed := false
		if userBucket.tokensCount >= 1.0 {
			userBucket.tokensCount -= 1.0
			isAllowed = true
		}

		storage.Unlock()

		if !isAllowed {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(serverContext, serverRequest)
	}
}
