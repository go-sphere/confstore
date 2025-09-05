package codec

import (
	"errors"
	"testing"
)

func TestStringCodec_UnmarshalNilPointer(t *testing.T) {
	c := StringCodec()
	var sp *string // nil *string
	if err := c.Unmarshal([]byte("abc"), sp); err == nil {
		t.Fatal("expected error for nil *string, got nil")
	} else if !errors.Is(err, ErrNilPointer) {
		t.Fatalf("expected ErrNilPointer, got %v", err)
	}
}
