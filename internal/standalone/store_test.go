package standalone

import (
	"database/sql"
	"errors"
	"path/filepath"
	"sort"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "konbu.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestOpen_AppliesSchemaAndIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "konbu.db")

	s1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	s2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer s2.Close()

	// default user row must exist exactly once after re-open
	var n int
	if err := s2.db.QueryRow(`SELECT COUNT(*) FROM users WHERE id = ?`, DefaultUserID).Scan(&n); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if n != 1 {
		t.Fatalf("default user row count = %d, want 1", n)
	}
}

func TestMemo_CRUD_WithTags(t *testing.T) {
	s := openTestStore(t)

	created, err := s.CreateMemo("hello", "world body", []string{"work", "draft"})
	if err != nil {
		t.Fatalf("CreateMemo: %v", err)
	}
	if created.ID == "" || created.Title != "hello" || created.Content != "world body" {
		t.Fatalf("CreateMemo returned %+v", created)
	}
	if names := tagNamesSorted(created.Tags); !equalStrings(names, []string{"draft", "work"}) {
		t.Fatalf("tags = %v, want [draft work]", names)
	}

	got, err := s.GetMemo(created.ID)
	if err != nil {
		t.Fatalf("GetMemo: %v", err)
	}
	if got.Title != "hello" {
		t.Fatalf("GetMemo title = %q", got.Title)
	}

	updated, err := s.UpdateMemo(created.ID, map[string]any{
		"title":   "hi",
		"content": "updated body",
		"tags":    []any{"work", "urgent"}, // drop "draft", add "urgent"
	})
	if err != nil {
		t.Fatalf("UpdateMemo: %v", err)
	}
	if updated.Title != "hi" || updated.Content != "updated body" {
		t.Fatalf("UpdateMemo returned %+v", updated)
	}
	if names := tagNamesSorted(updated.Tags); !equalStrings(names, []string{"urgent", "work"}) {
		t.Fatalf("post-update tags = %v, want [urgent work]", names)
	}

	if err := s.DeleteMemo(created.ID); err != nil {
		t.Fatalf("DeleteMemo: %v", err)
	}
	if _, err := s.GetMemo(created.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetMemo after delete: err = %v, want sql.ErrNoRows", err)
	}
	memos, err := s.ListMemos()
	if err != nil {
		t.Fatalf("ListMemos: %v", err)
	}
	for _, m := range memos {
		if m.ID == created.ID {
			t.Fatalf("ListMemos still returns soft-deleted memo %s", m.ID)
		}
	}
}

func TestTodo_StatusAndDueDateLifecycle(t *testing.T) {
	s := openTestStore(t)

	if _, err := s.CreateTodo(map[string]any{"title": "  "}); err == nil {
		t.Fatalf("CreateTodo with blank title should fail")
	}

	created, err := s.CreateTodo(map[string]any{
		"title":    "write tests",
		"due_date": "2026-05-20",
	})
	if err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}
	if created.Status != "open" {
		t.Fatalf("default status = %q, want open", created.Status)
	}
	if created.DueDate != "2026-05-20" {
		t.Fatalf("DueDate = %q", created.DueDate)
	}

	if err := s.SetTodoStatus(created.ID, "done"); err != nil {
		t.Fatalf("SetTodoStatus: %v", err)
	}
	got, err := s.GetTodo(created.ID)
	if err != nil {
		t.Fatalf("GetTodo: %v", err)
	}
	if got.Status != "done" {
		t.Fatalf("status after SetTodoStatus = %q", got.Status)
	}

	// passing an explicit empty due_date should clear it
	cleared, err := s.UpdateTodo(created.ID, map[string]any{"due_date": ""})
	if err != nil {
		t.Fatalf("UpdateTodo clear due_date: %v", err)
	}
	if cleared.DueDate != "" {
		t.Fatalf("DueDate after clear = %q, want empty", cleared.DueDate)
	}

	if err := s.SetTodoStatus("does-not-exist", "done"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("SetTodoStatus missing id: err = %v, want sql.ErrNoRows", err)
	}
}

func TestEvent_AllDayAndUpdate(t *testing.T) {
	s := openTestStore(t)

	if _, err := s.CreateEvent(map[string]any{"title": "no start"}); err == nil {
		t.Fatalf("CreateEvent without start_at should fail")
	}

	ev, err := s.CreateEvent(map[string]any{
		"title":    "standup",
		"start_at": "2026-05-16T10:00:00Z",
		"end_at":   "2026-05-16T10:30:00Z",
	})
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}
	if ev.AllDay {
		t.Fatalf("AllDay should default to false")
	}
	if ev.EndAt != "2026-05-16T10:30:00Z" {
		t.Fatalf("EndAt = %q", ev.EndAt)
	}

	updated, err := s.UpdateEvent(ev.ID, map[string]any{
		"all_day": true,
		"end_at":  "",
	})
	if err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}
	if !updated.AllDay {
		t.Fatalf("AllDay should be true after update")
	}
	if updated.EndAt != "" {
		t.Fatalf("EndAt after clear = %q, want empty", updated.EndAt)
	}
}

func TestTags_NormalizedAcrossResources(t *testing.T) {
	s := openTestStore(t)

	if _, err := s.CreateMemo("m1", "", []string{"shared"}); err != nil {
		t.Fatalf("CreateMemo: %v", err)
	}
	if _, err := s.CreateTodo(map[string]any{"title": "t1", "tags": []any{"shared"}}); err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}

	var n int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM tags WHERE user_id = ? AND name = ?`,
		s.userID, "shared",
	).Scan(&n); err != nil {
		t.Fatalf("count tags: %v", err)
	}
	if n != 1 {
		t.Fatalf("tag rows = %d, want 1 (upsert should dedupe)", n)
	}
}

func TestSearch_AcrossKindsAndExcludesDeleted(t *testing.T) {
	s := openTestStore(t)

	memo, err := s.CreateMemo("alpha memo", "the body mentions sphinx once", nil)
	if err != nil {
		t.Fatalf("CreateMemo: %v", err)
	}
	if _, err := s.CreateTodo(map[string]any{"title": "todo about sphinx", "description": ""}); err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}
	if _, err := s.CreateEvent(map[string]any{
		"title":       "weekly sphinx review",
		"start_at":    "2026-05-16T09:00:00Z",
		"description": "",
	}); err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}

	results, err := s.Search("Sphinx") // case-insensitive
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	counts := map[string]int{}
	for _, r := range results {
		counts[r.Type]++
	}
	if counts["memo"] != 1 || counts["todo"] != 1 || counts["event"] != 1 {
		t.Fatalf("counts = %v, want one of each kind", counts)
	}

	if err := s.DeleteMemo(memo.ID); err != nil {
		t.Fatalf("DeleteMemo: %v", err)
	}
	results, err = s.Search("sphinx")
	if err != nil {
		t.Fatalf("Search after delete: %v", err)
	}
	for _, r := range results {
		if r.Type == "memo" {
			t.Fatalf("soft-deleted memo should not appear in search results: %+v", r)
		}
	}

	if results, err := s.Search("   "); err != nil || results != nil {
		t.Fatalf("Search(blank) = (%v, %v), want (nil, nil)", results, err)
	}
}

func tagNamesSorted(tags []Tag) []string {
	out := make([]string, len(tags))
	for i, t := range tags {
		out[i] = t.Name
	}
	sort.Strings(out)
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
