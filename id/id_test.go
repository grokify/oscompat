package id_test

import (
	"sync"
	"testing"

	"github.com/grokify/oscompat/id"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		byteLen int
		wantLen int
	}{
		{"8 bytes", 8, 16},
		{"16 bytes", 16, 32},
		{"32 bytes", 32, 64},
		{"1 byte", 1, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := id.Generate(tt.byteLen)
			if len(got) != tt.wantLen {
				t.Errorf("Generate(%d) returned length %d, want %d", tt.byteLen, len(got), tt.wantLen)
			}
		})
	}
}

func TestGenerate16(t *testing.T) {
	got := id.Generate16()
	if len(got) != 16 {
		t.Errorf("Generate16() returned length %d, want 16", len(got))
	}
}

func TestGenerate32(t *testing.T) {
	got := id.Generate32()
	if len(got) != 32 {
		t.Errorf("Generate32() returned length %d, want 32", len(got))
	}
}

func TestGenerateUniqueness(t *testing.T) {
	// Generate many IDs and verify they're all unique.
	// This is the key test for Windows compatibility - on Windows with
	// time-based generation, rapid calls would produce duplicates.
	const count = 10000
	ids := make(map[string]struct{}, count)

	for i := 0; i < count; i++ {
		id := id.Generate16()
		if _, exists := ids[id]; exists {
			t.Fatalf("duplicate ID generated at iteration %d: %s", i, id)
		}
		ids[id] = struct{}{}
	}
}

func TestGenerateUniquenessConcurrent(t *testing.T) {
	// Test uniqueness under concurrent access.
	const (
		goroutines = 100
		perRoutine = 100
	)

	var (
		mu  sync.Mutex
		ids = make(map[string]struct{}, goroutines*perRoutine)
		wg  sync.WaitGroup
	)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localIDs := make([]string, perRoutine)
			for j := 0; j < perRoutine; j++ {
				localIDs[j] = id.Generate16()
			}

			mu.Lock()
			defer mu.Unlock()
			for _, id := range localIDs {
				if _, exists := ids[id]; exists {
					t.Errorf("duplicate ID generated: %s", id)
					return
				}
				ids[id] = struct{}{}
			}
		}()
	}

	wg.Wait()
}

func TestGenerateHexFormat(t *testing.T) {
	// Verify output is valid hex.
	got := id.Generate(8)
	for _, c := range got {
		isDigit := c >= '0' && c <= '9'
		isHexLower := c >= 'a' && c <= 'f'
		if !isDigit && !isHexLower {
			t.Errorf("Generate() returned non-hex character: %c in %s", c, got)
		}
	}
}

func BenchmarkGenerate8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		id.Generate(8)
	}
}

func BenchmarkGenerate16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		id.Generate16()
	}
}

func BenchmarkGenerate32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		id.Generate32()
	}
}
