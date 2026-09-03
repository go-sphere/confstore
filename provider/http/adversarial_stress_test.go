package http

import (
	"context"
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestHTTPAdversarialConcurrentQueriesCustomTimeout stress tests concurrent queries
// against an HTTP provider configured with a custom client and custom timeout.
// It verifies:
// 1. WithClient preserves the caller's custom timeout (WithTimeout is ignored).
// 2. Bounded concurrent goroutines querying endpoints experience no data races.
// 3. Fast queries succeed reliably; slow queries consistently time out according to custom client timeout.
func TestHTTPAdversarialConcurrentQueriesCustomTimeout(t *testing.T) {
	ts := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		switch r.URL.Path {
		case "/fast":
			w.WriteHeader(nethttp.StatusOK)
			_, _ = w.Write([]byte(`{"app":"sphere","env":"prod"}`))
		case "/slow":
			// Sleep longer than custom client timeout (500ms > 200ms)
			select {
			case <-time.After(500 * time.Millisecond):
				w.WriteHeader(nethttp.StatusOK)
				_, _ = w.Write([]byte(`{"delayed":true}`))
			case <-r.Context().Done():
				return
			}
		default:
			nethttp.NotFound(w, r)
		}
	}))
	defer ts.Close()

	customTimeout := 200 * time.Millisecond
	customClient := &nethttp.Client{
		Timeout: customTimeout,
		Transport: &nethttp.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
		},
	}

	// Supply WithTimeout(10*time.Second) alongside WithClient;
	// contract specifies WithClient takes precedence and must NOT be mutated.
	providerFast := New(ts.URL+"/fast", WithClient(customClient), WithTimeout(10*time.Second))
	providerSlow := New(ts.URL+"/slow", WithClient(customClient), WithTimeout(10*time.Second))

	if customClient.Timeout != customTimeout {
		t.Fatalf("custom client Timeout was mutated: got %v, want %v", customClient.Timeout, customTimeout)
	}

	const concurrency = 16
	const queriesPerWorker = 5
	var wg sync.WaitGroup

	var fastSuccessCount int64
	var slowTimeoutCount int64
	errCh := make(chan error, concurrency*queriesPerWorker*2)

	wg.Add(concurrency)
	for w := 0; w < concurrency; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < queriesPerWorker; i++ {
				// 1. Query fast endpoint
				data, err := providerFast.Read(context.Background())
				if err != nil {
					errCh <- err
					return
				}
				if !strings.Contains(string(data), `"env":"prod"`) {
					errCh <- errors.New("fast query payload mismatch: " + string(data))
					return
				}
				atomic.AddInt64(&fastSuccessCount, 1)

				// 2. Query slow endpoint (should timeout via custom client)
				_, err = providerSlow.Read(context.Background())
				if err == nil {
					errCh <- errors.New("slow query unexpectedly succeeded despite custom client timeout")
					return
				}
				atomic.AddInt64(&slowTimeoutCount, 1)
			}
		}(w)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent HTTP provider error: %v", err)
	}

	expectedCount := int64(concurrency * queriesPerWorker)
	if fastSuccessCount != expectedCount {
		t.Errorf("fast success count mismatch: got %d, want %d", fastSuccessCount, expectedCount)
	}
	if slowTimeoutCount != expectedCount {
		t.Errorf("slow timeout count mismatch: got %d, want %d", slowTimeoutCount, expectedCount)
	}

	// Verify custom client timeout remained untouched after concurrent requests
	if customClient.Timeout != customTimeout {
		t.Errorf("custom client Timeout was modified during execution: got %v, want %v", customClient.Timeout, customTimeout)
	}
}
