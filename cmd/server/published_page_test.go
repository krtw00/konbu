package main

import (
	"strings"
	"testing"
)

func TestSummarizeMemoContentStripsMarkdownAndTruncates(t *testing.T) {
	content := "# Heading\n\nThis is **markdown** with [a link](https://example.com) and enough content to exercise the summary helper without keeping raw markdown syntax in the result."
	got := summarizeMemoContent(&content)

	if strings.Contains(got, "**") || strings.Contains(got, "[") || strings.Contains(got, "]") {
		t.Fatalf("expected markdown syntax to be stripped, got %q", got)
	}
	if len(got) > 160 {
		t.Fatalf("expected summary to be truncated to 160 chars, got %d", len(got))
	}
}

func TestInjectPageMetadataAddsSEOAndOGPTags(t *testing.T) {
	indexHTML := "<html><head><title>konbu</title></head><body></body></html>"
	got := injectPageMetadata(indexHTML, pageMeta{
		Title:       "Published Memo",
		Description: "Memo description",
		URL:         "https://example.com/memo/published-memo",
		ImageURL:    "https://example.com/hero.png",
	})

	for _, needle := range []string{
		"<title>Published Memo | konbu</title>",
		`<meta name="description" content="Memo description" />`,
		`<meta property="og:title" content="Published Memo" />`,
		`<meta property="og:url" content="https://example.com/memo/published-memo" />`,
		`<meta name="twitter:card" content="summary_large_image" />`,
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("expected injected HTML to contain %q", needle)
		}
	}
}
