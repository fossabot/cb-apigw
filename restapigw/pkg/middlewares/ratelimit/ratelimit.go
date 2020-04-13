// Package ratelimit - Token Bucket 기반의 Rate Limit 처리를 지원하는 패키지
package ratelimit

import (
	"context"
	"errors"
	"time"

	"github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/ratelimit/backend"
	"github.com/cloud-barista/cb-apigw/restapigw/pkg/middlewares/ratelimit/limiter"
)

// ===== [ Constants and Variables ] =====

var (
	// ErrLimited - Rate Limit 제한이 되었을 때의 오류
	ErrLimited = errors.New("ERROR: rate limit exceeded")
)

// ===== [ Types ] =====

// ILimiter - Rate limit 운영을 위한 인터페이스
type ILimiter interface {
	// Rate limit 초과 여부 검증
	Allow() bool
}

// LimiterStore - 지정한 키에 해당하는 Limiter 정보를 검증하는 함수 구조
type LimiterStore func(string) ILimiter

// RateLimiter - Rate limit 운영을 위한 Bucket Wrapper 구조
type RateLimiter struct {
	limiter *limiter.Bucket
}

// ===== [ Implementations ] =====

// Allow - Rate Limit 처리를 위해 Bucket에서 Token 사용이 가능한지를 검증하고, 1개의 Token을 사용한다.
func (rl RateLimiter) Allow() bool {
	return rl.limiter.TakeAvailable(1) > 0
}

// ===== [ Private Functions ] =====
// ===== [ Public Functions ] =====

// NewLimitterStore - Rate limit 정보 저장을 위한 Limiter Store 생성
func NewLimitterStore(maxRate float64, capacity int64, backend backend.IBackend) LimiterStore {
	f := func() interface{} {
		return NewLimiterWithRate(maxRate, capacity)
	}
	return func(t string) ILimiter {
		return backend.Load(t, f).(RateLimiter)
	}
}

// NewMemoryStore - Memory를 기반으로 하는 LimiterStore 생성
func NewMemoryStore(maxRate float64, capacity int64) LimiterStore {
	return NewLimitterStore(maxRate, capacity, backend.DefaultShardedMemoryBackend(context.Background()))
}

// NewLimiterWithRate - 지정한 비율과 최대 용량을 기준으로 Rate Limiter 생성
func NewLimiterWithRate(maxRate float64, capacity int64) RateLimiter {
	return RateLimiter{
		limiter.NewBucketWithRate(maxRate, capacity),
	}
}

// NewLimiterWithFillInterval - 지정한 기간과 최대 용량을 기준으로 Rate Limiter 생성
func NewLimiterWithFillInterval(fillInterval time.Duration, capacity int64) RateLimiter {
	return RateLimiter{
		limiter.NewBucket(fillInterval, capacity),
	}
}
