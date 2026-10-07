package utils

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// WaitForHTTP waits for an HTTP endpoint to be ready
func WaitForHTTP(url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for %s to be ready", url)
		case <-ticker.C:
			resp, err := http.Get(url)
			if err == nil && resp.StatusCode < 500 {
				resp.Body.Close()
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}

// WaitForGRPC waits for a gRPC endpoint to be ready
// For now, we'll use a simple sleep or can extend with actual gRPC health check
func WaitForGRPC(timeout time.Duration) error {
	time.Sleep(timeout)
	return nil
}
