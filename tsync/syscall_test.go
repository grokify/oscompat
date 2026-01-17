//go:build !windows

package tsync_test

import (
	"syscall"
	"testing"
	"time"

	"github.com/grokify/oscompat/tsync"
)

func TestFromTimespec(t *testing.T) {
	// Test with known values
	tests := []struct {
		name string
		ts   syscall.Timespec
		want time.Time
	}{
		{
			"epoch",
			syscall.Timespec{Sec: 0, Nsec: 0},
			time.Unix(0, 0),
		},
		{
			"with seconds",
			syscall.Timespec{Sec: 1234567890, Nsec: 0},
			time.Unix(1234567890, 0),
		},
		{
			"with nanoseconds",
			syscall.Timespec{Sec: 1234567890, Nsec: 123456789},
			time.Unix(1234567890, 123456789),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tsync.FromTimespec(tt.ts)
			if !got.Equal(tt.want) {
				t.Errorf("FromTimespec(%+v) = %v, want %v", tt.ts, got, tt.want)
			}
		})
	}
}

func TestFromTimeval(t *testing.T) {
	// Test with known values
	tests := []struct {
		name string
		tv   syscall.Timeval
		want time.Time
	}{
		{
			"epoch",
			syscall.Timeval{Sec: 0, Usec: 0},
			time.Unix(0, 0),
		},
		{
			"with seconds",
			syscall.Timeval{Sec: 1234567890, Usec: 0},
			time.Unix(1234567890, 0),
		},
		{
			"with microseconds",
			syscall.Timeval{Sec: 1234567890, Usec: 123456},
			time.Unix(1234567890, 123456000), // usec * 1000 = nsec
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tsync.FromTimeval(tt.tv)
			if !got.Equal(tt.want) {
				t.Errorf("FromTimeval(%+v) = %v, want %v", tt.tv, got, tt.want)
			}
		})
	}
}
