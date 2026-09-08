package reader

import (
	"bytes"
	"context"
	"math/rand"
	"sync"
	"testing"
)

// TestBytesAdversarialConcurrentReadsAndMutations stress tests the Bytes provider
// with 100 concurrent goroutines continuously reading and mutating their returned slices,
// verifying zero data races and zero data corruption across callers.
func TestBytesAdversarialConcurrentReadsAndMutations(t *testing.T) {
	orig := []byte("immutable-confstore-payload-adversarial-verification-9876543210")
	p := NewBytes(orig)

	const goroutines = 100
	const iterations = 100
	var wg sync.WaitGroup
	errCh := make(chan error, goroutines*iterations)

	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(workerID int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(int64(workerID)))
			for i := 0; i < iterations; i++ {
				data, err := p.Read(context.Background())
				if err != nil {
					t.Errorf("worker %d iteration %d read failed: %v", workerID, i, err)
					return
				}
				if !bytes.Equal(data, orig) {
					t.Errorf("worker %d iteration %d corrupted read: got %q, want %q", workerID, i, string(data), string(orig))
					return
				}

				// Adversarially scramble the returned slice in-place
				for idx := range data {
					data[idx] = byte(r.Intn(256))
				}

				// Second read to verify immutability after local mutation
				data2, err := p.Read(context.Background())
				if err != nil {
					t.Errorf("worker %d iteration %d second read failed: %v", workerID, i, err)
					return
				}
				if !bytes.Equal(data2, orig) {
					t.Errorf("worker %d iteration %d corrupted read after mutation: got %q, want %q", workerID, i, string(data2), string(orig))
					return
				}
			}
		}(g)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent mutation test error: %v", err)
	}
}

// TestBytesConstructorSliceAliasing checks if mutating the slice originally passed to NewBytes
// affects subsequent reads.
func TestBytesConstructorSliceAliasing(t *testing.T) {
	src := []byte("initial-data")
	p := NewBytes(src)

	// NewBytes clones the input, so mutating src afterwards must not affect reads.
	src[0] = 'M'
	got, _ := p.Read(context.Background())
	if string(got) != "initial-data" {
		t.Fatalf("NewBytes must clone constructor input: got %q, want %q", string(got), "initial-data")
	}
}
