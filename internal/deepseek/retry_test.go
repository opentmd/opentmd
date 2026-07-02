package deepseek

import (
	"testing"
	"time"
)

func TestIsRetryableStatus(t *testing.T) {
	if !isRetryableStatus(429) {
		t.Fatal("429 should retry")
	}
	if isRetryableStatus(400) {
		t.Fatal("400 should not retry")
	}
}

func TestBackoffWithinMax(t *testing.T) {
	p := DefaultRetryPolicy()
	d := backoff(5, p)
	if d > p.MaxDelay+time.Second {
		t.Fatalf("backoff too large: %v", d)
	}
}
