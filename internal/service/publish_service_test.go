package service

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNormalizePublishSlug(t *testing.T) {
	if got := normalizePublishSlug(" Hello, Konbu World! "); got != "hello-konbu-world" {
		t.Fatalf("unexpected slug: %q", got)
	}
	if got := normalizePublishSlug("___multi   space___"); got != "multi-space" {
		t.Fatalf("unexpected slug normalization: %q", got)
	}
}

func TestFallbackPublishSlugUsesResourcePrefixWhenTitleCannotBeNormalized(t *testing.T) {
	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	got := fallbackPublishSlug(PublishedResourceMemo, id, "配信予定")
	if !strings.HasPrefix(got, "memo-11111111") {
		t.Fatalf("unexpected fallback slug: %q", got)
	}
}

func TestNormalizePublishVisibility(t *testing.T) {
	if got := normalizePublishVisibility(" PUBLIC "); got != PublishVisibilityPublic {
		t.Fatalf("unexpected visibility: %q", got)
	}
	if got := normalizePublishVisibility("invalid"); got != "" {
		t.Fatalf("expected empty visibility for invalid input, got %q", got)
	}
}
