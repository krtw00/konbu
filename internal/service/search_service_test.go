package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/model"
)

func TestMakeTypeSetDefaults(t *testing.T) {
	got := makeTypeSet(nil)

	for _, key := range []string{"memo", "todo", "event", "tool"} {
		if !got[key] {
			t.Fatalf("expected default type %q to be enabled", key)
		}
	}
}

func TestMakeTypeSetExplicit(t *testing.T) {
	got := makeTypeSet([]string{"memo", "tool"})

	if len(got) != 2 || !got["memo"] || !got["tool"] {
		t.Fatalf("unexpected type set: %#v", got)
	}
}

func TestPaginateBounds(t *testing.T) {
	now := time.Now()
	items := []model.SearchResult{
		{ID: uuid.New(), UpdatedAt: now},
		{ID: uuid.New(), UpdatedAt: now},
		{ID: uuid.New(), UpdatedAt: now},
	}

	got := paginate(items, 1, 5)
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}

	got = paginate(items, 10, 5)
	if len(got) != 0 {
		t.Fatalf("expected empty page, got %d items", len(got))
	}
}

func TestTruncateRuneAware(t *testing.T) {
	got := truncate("こんにちは世界", 4)
	want := "こんにち..."
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
