package service

import (
	"testing"
	"time"

	"github.com/krtw00/konbu/internal/repository"
)

func TestToolNeedsIconRefresh(t *testing.T) {
	now := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)
	stale := now.Add(-25 * time.Hour)
	fresh := now.Add(-23 * time.Hour)
	exact := now.Add(-24 * time.Hour)

	tests := []struct {
		name string
		tool repository.Tool
		want bool
	}{
		{
			name: "empty url is ignored",
			tool: repository.Tool{},
			want: false,
		},
		{
			name: "nil checked time refreshes",
			tool: repository.Tool{URL: "https://example.com"},
			want: true,
		},
		{
			name: "stale checked time refreshes",
			tool: repository.Tool{URL: "https://example.com", IconCheckedAt: &stale},
			want: true,
		},
		{
			name: "exactly one day old refreshes",
			tool: repository.Tool{URL: "https://example.com", IconCheckedAt: &exact},
			want: true,
		},
		{
			name: "recent check is skipped",
			tool: repository.Tool{URL: "https://example.com", IconCheckedAt: &fresh},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toolNeedsIconRefresh(tt.tool, now); got != tt.want {
				t.Fatalf("toolNeedsIconRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}
