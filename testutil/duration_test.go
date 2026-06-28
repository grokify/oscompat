package testutil

import (
	"testing"
	"time"
)

func TestDurationNonNegative(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     bool
	}{
		{"positive", time.Second, true},
		{"zero", 0, true},
		{"negative", -time.Second, false},
		{"small positive", time.Nanosecond, true},
		{"small negative", -time.Nanosecond, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DurationNonNegative(tt.duration)
			if got != tt.want {
				t.Errorf("DurationNonNegative(%v) = %v, want %v", tt.duration, got, tt.want)
			}
		})
	}
}

func TestDurationPositiveOrZero(t *testing.T) {
	// Verify it's an alias for DurationNonNegative
	if DurationPositiveOrZero(0) != DurationNonNegative(0) {
		t.Error("DurationPositiveOrZero should be an alias for DurationNonNegative")
	}
	if DurationPositiveOrZero(time.Second) != DurationNonNegative(time.Second) {
		t.Error("DurationPositiveOrZero should be an alias for DurationNonNegative")
	}
	if DurationPositiveOrZero(-time.Second) != DurationNonNegative(-time.Second) {
		t.Error("DurationPositiveOrZero should be an alias for DurationNonNegative")
	}
}

func TestAssertDurationNonNegative(t *testing.T) {
	// Test with non-negative values (should pass)
	t.Run("positive", func(t *testing.T) {
		AssertDurationNonNegative(t, time.Second)
	})

	t.Run("zero", func(t *testing.T) {
		AssertDurationNonNegative(t, 0)
	})
}
