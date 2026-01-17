package tsync_test

import (
	"testing"
	"time"

	"github.com/grokify/oscompat/tsync"
)

func TestTolerance(t *testing.T) {
	tolerance := tsync.Tolerance()
	if tolerance != time.Second {
		t.Errorf("Tolerance() = %v, want %v", tolerance, time.Second)
	}
}

func TestEqual(t *testing.T) {
	base := time.Now()

	tests := []struct {
		name string
		t1   time.Time
		t2   time.Time
		want bool
	}{
		{"same time", base, base, true},
		{"half second apart", base, base.Add(500 * time.Millisecond), true},
		{"one second apart", base, base.Add(time.Second), true},
		{"1.5 seconds apart", base, base.Add(1500 * time.Millisecond), false},
		{"two seconds apart", base, base.Add(2 * time.Second), false},
		{"negative half second", base, base.Add(-500 * time.Millisecond), true},
		{"negative two seconds", base, base.Add(-2 * time.Second), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tsync.Equal(tt.t1, tt.t2)
			if got != tt.want {
				t.Errorf("Equal(%v, %v) = %v, want %v", tt.t1, tt.t2, got, tt.want)
			}
		})
	}
}

func TestEqualWithTolerance(t *testing.T) {
	base := time.Now()

	tests := []struct {
		name      string
		t1        time.Time
		t2        time.Time
		tolerance time.Duration
		want      bool
	}{
		{"within tolerance", base, base.Add(100 * time.Millisecond), 200 * time.Millisecond, true},
		{"at tolerance boundary", base, base.Add(200 * time.Millisecond), 200 * time.Millisecond, true},
		{"beyond tolerance", base, base.Add(300 * time.Millisecond), 200 * time.Millisecond, false},
		{"FAT32 tolerance", base, base.Add(1500 * time.Millisecond), tsync.FAT32Tolerance, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tsync.EqualWithTolerance(tt.t1, tt.t2, tt.tolerance)
			if got != tt.want {
				t.Errorf("EqualWithTolerance(%v, %v, %v) = %v, want %v",
					tt.t1, tt.t2, tt.tolerance, got, tt.want)
			}
		})
	}
}

func TestBefore(t *testing.T) {
	base := time.Now()

	tests := []struct {
		name string
		t1   time.Time
		t2   time.Time
		want bool
	}{
		{"same time", base, base, false},
		{"within tolerance", base, base.Add(500 * time.Millisecond), false},
		{"at tolerance", base, base.Add(time.Second), false},
		{"beyond tolerance", base, base.Add(2 * time.Second), true},
		{"t1 after t2", base.Add(2 * time.Second), base, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tsync.Before(tt.t1, tt.t2)
			if got != tt.want {
				t.Errorf("Before(%v, %v) = %v, want %v", tt.t1, tt.t2, got, tt.want)
			}
		})
	}
}

func TestAfter(t *testing.T) {
	base := time.Now()

	tests := []struct {
		name string
		t1   time.Time
		t2   time.Time
		want bool
	}{
		{"same time", base, base, false},
		{"within tolerance", base.Add(500 * time.Millisecond), base, false},
		{"at tolerance", base.Add(time.Second), base, false},
		{"beyond tolerance", base.Add(2 * time.Second), base, true},
		{"t1 before t2", base, base.Add(2 * time.Second), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tsync.After(tt.t1, tt.t2)
			if got != tt.want {
				t.Errorf("After(%v, %v) = %v, want %v", tt.t1, tt.t2, got, tt.want)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	base := time.Now()

	tests := []struct {
		name string
		t1   time.Time
		t2   time.Time
		want int
	}{
		{"same time", base, base, 0},
		{"within tolerance", base, base.Add(500 * time.Millisecond), 0},
		{"t1 before t2", base, base.Add(2 * time.Second), -1},
		{"t1 after t2", base.Add(2 * time.Second), base, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tsync.Compare(tt.t1, tt.t2)
			if got != tt.want {
				t.Errorf("Compare(%v, %v) = %v, want %v", tt.t1, tt.t2, got, tt.want)
			}
		})
	}
}

func TestNewer(t *testing.T) {
	base := time.Now()
	later := base.Add(2 * time.Second)

	// Clear winner
	if got := tsync.Newer(base, later); !got.Equal(later) {
		t.Errorf("Newer() = %v, want %v", got, later)
	}
	if got := tsync.Newer(later, base); !got.Equal(later) {
		t.Errorf("Newer() = %v, want %v", got, later)
	}

	// Within tolerance - should return t1
	almostSame := base.Add(500 * time.Millisecond)
	if got := tsync.Newer(base, almostSame); !got.Equal(base) {
		t.Errorf("Newer() within tolerance = %v, want %v", got, base)
	}
}

func TestOlder(t *testing.T) {
	base := time.Now()
	later := base.Add(2 * time.Second)

	// Clear winner
	if got := tsync.Older(base, later); !got.Equal(base) {
		t.Errorf("Older() = %v, want %v", got, base)
	}
	if got := tsync.Older(later, base); !got.Equal(base) {
		t.Errorf("Older() = %v, want %v", got, base)
	}

	// Within tolerance - should return t1
	almostSame := base.Add(500 * time.Millisecond)
	if got := tsync.Older(base, almostSame); !got.Equal(base) {
		t.Errorf("Older() within tolerance = %v, want %v", got, base)
	}
}

func TestTruncate(t *testing.T) {
	// Use a time with nanosecond precision
	base := time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC)

	tests := []struct {
		name      string
		precision time.Duration
		wantNano  int
	}{
		{"to second", time.Second, 0},
		{"to millisecond", time.Millisecond, 123000000},
		{"to microsecond", time.Microsecond, 123456000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tsync.Truncate(base, tt.precision)
			if got.Nanosecond() != tt.wantNano {
				t.Errorf("Truncate() nanosecond = %v, want %v", got.Nanosecond(), tt.wantNano)
			}
		})
	}
}

func TestTruncateToSecond(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC)
	got := tsync.TruncateToSecond(base)

	if got.Nanosecond() != 0 {
		t.Errorf("TruncateToSecond() nanosecond = %v, want 0", got.Nanosecond())
	}
	if got.Second() != 45 {
		t.Errorf("TruncateToSecond() second = %v, want 45", got.Second())
	}
}

func TestConstants(t *testing.T) {
	if tsync.DefaultTolerance != time.Second {
		t.Errorf("DefaultTolerance = %v, want %v", tsync.DefaultTolerance, time.Second)
	}
	if tsync.FAT32Tolerance != 2*time.Second {
		t.Errorf("FAT32Tolerance = %v, want %v", tsync.FAT32Tolerance, 2*time.Second)
	}
	if tsync.HighPrecisionTolerance != 100*time.Millisecond {
		t.Errorf("HighPrecisionTolerance = %v, want %v", tsync.HighPrecisionTolerance, 100*time.Millisecond)
	}
}
