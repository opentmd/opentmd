package deepseek

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/opentmd/opentmd/internal/llm"
)

type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    8 * time.Second,
	}
}

func isRetryableStatus(code int) bool {
	switch code {
	case 408, 425, 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

func backoff(attempt int, policy RetryPolicy) time.Duration {
	exp := policy.BaseDelay << attempt
	if exp > policy.MaxDelay {
		exp = policy.MaxDelay
	}
	jitter := time.Duration(rand.Int63n(int64(exp / 4)))
	return exp - exp/8 + jitter
}

func (c *Client) CompleteWithRetry(ctx context.Context, req llm.ChatRequest, policy RetryPolicy) (*llm.ChatResponse, error) {
	if policy.MaxAttempts <= 0 {
		policy = DefaultRetryPolicy()
	}
	var lastErr error
	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		resp, err := c.completeOnce(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		retryable := false
		var apiErr *apiStatusError
		if ok := asAPIStatus(err, &apiErr); ok {
			retryable = isRetryableStatus(apiErr.Code)
		}
		if !retryable {
			return nil, err
		}
		if attempt+1 >= policy.MaxAttempts {
			break
		}
		delay := backoff(attempt, policy)
		if apiErr != nil && apiErr.RetryAfter > 0 {
			delay = apiErr.RetryAfter
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil, fmt.Errorf("request failed after %d attempts: %w", policy.MaxAttempts, lastErr)
}

type apiStatusError struct {
	Code       int
	Body       string
	RetryAfter time.Duration
}

func (e *apiStatusError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.Code, e.Body)
}

func asAPIStatus(err error, target **apiStatusError) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*apiStatusError); ok {
		*target = e
		return true
	}
	return false
}

func parseRetryAfter(h http.Header) time.Duration {
	v := h.Get("Retry-After")
	if v == "" {
		return 0
	}
	if secs, err := time.ParseDuration(v + "s"); err == nil {
		return secs
	}
	return 0
}
