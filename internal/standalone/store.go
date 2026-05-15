// Package standalone provides a self-contained SQLite-backed store used by
// `konbu mcp --standalone`. It is intentionally decoupled from the PostgreSQL
// repositories under internal/repository so that the standalone build stays
// dependency-light and free of Postgres-isms (now(), JSONB, RETURNING, pg_trgm).
package standalone

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	_ "modernc.org/sqlite"
)

// DefaultUserID is the fixed user attributed to every standalone-created row.
// Standalone mode is a single-user local context; we mint one UUID up-front
// instead of tracking auth.
const DefaultUserID = "00000000-0000-0000-0000-000000000001"

const defaultUserEmail = "default@local"

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id          TEXT PRIMARY KEY,
    email       TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tags (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    name        TEXT NOT NULL,
    created_at  TEXT NOT NULL,
    UNIQUE(user_id, name)
);

CREATE TABLE IF NOT EXISTS memos (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    title       TEXT NOT NULL,
    content     TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    deleted_at  TEXT
);
CREATE INDEX IF NOT EXISTS idx_memos_user_updated ON memos(user_id, updated_at);

CREATE TABLE IF NOT EXISTS memo_tags (
    memo_id  TEXT NOT NULL,
    tag_id   TEXT NOT NULL,
    PRIMARY KEY (memo_id, tag_id)
);

CREATE TABLE IF NOT EXISTS todos (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'open',
    due_date     TEXT,
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL,
    deleted_at   TEXT
);
CREATE INDEX IF NOT EXISTS idx_todos_user_created ON todos(user_id, created_at);

CREATE TABLE IF NOT EXISTS todo_tags (
    todo_id  TEXT NOT NULL,
    tag_id   TEXT NOT NULL,
    PRIMARY KEY (todo_id, tag_id)
);

CREATE TABLE IF NOT EXISTS events (
    id              TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    start_at        TEXT NOT NULL,
    end_at          TEXT,
    all_day         INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL,
    deleted_at      TEXT
);
CREATE INDEX IF NOT EXISTS idx_events_user_start ON events(user_id, start_at);

CREATE TABLE IF NOT EXISTS event_tags (
    event_id  TEXT NOT NULL,
    tag_id    TEXT NOT NULL,
    PRIMARY KEY (event_id, tag_id)
);
`

// JSON shapes intentionally mirror internal/client types so the MCP server
// emits the same payload whether running against an HTTP backend or locally.

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Memo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Tags      []Tag  `json:"tags"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Todo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	DueDate     string `json:"due_date"`
	Tags        []Tag  `json:"tags"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Event struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartAt     string `json:"start_at"`
	EndAt       string `json:"end_at,omitempty"`
	AllDay      bool   `json:"all_day"`
	Tags        []Tag  `json:"tags"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type SearchResult struct {
	Type    string   `json:"type"`
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Snippet string   `json:"snippet"`
	Tags    []string `json:"tags"`
}

type Store struct {
	db     *sql.DB
	userID string
}

// DefaultDBPath returns ~/.konbu/konbu.db, the canonical local store location.
func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".konbu", "konbu.db"), nil
}

// Open prepares the SQLite database at path: creates parent dirs, applies the
// schema, and ensures the default user row exists. Pass an empty path to use
// the default location.
func Open(path string) (*Store, error) {
	if path == "" {
		var err error
		path, err = DefaultDBPath()
		if err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	s := &Store{db: db, userID: DefaultUserID}
	if err := s.ensureDefaultUser(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure default user: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) ensureDefaultUser() error {
	now := nowISO()
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO users (id, email, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		s.userID, defaultUserEmail, "default", now, now,
	)
	return err
}

func nowISO() string { return time.Now().UTC().Format(time.RFC3339) }

func newID() string { return uuid.NewString() }

// --- tags ---

// upsertTag returns the tag ID for (userID, name), creating the row when needed.
func (s *Store) upsertTag(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("empty tag name")
	}
	var id string
	err := s.db.QueryRow(`SELECT id FROM tags WHERE user_id = ? AND name = ?`, s.userID, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	id = newID()
	if _, err := s.db.Exec(
		`INSERT INTO tags (id, user_id, name, created_at) VALUES (?, ?, ?, ?)`,
		id, s.userID, name, nowISO(),
	); err != nil {
		return "", err
	}
	return id, nil
}

// setTags replaces the tag set on a resource. table is one of "memo_tags",
// "todo_tags", "event_tags"; refCol is the corresponding FK column.
func (s *Store) setTags(table, refCol, refID string, names []string) error {
	if _, err := s.db.Exec(fmt.Sprintf(`DELETE FROM %s WHERE %s = ?`, table, refCol), refID); err != nil {
		return err
	}
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		tagID, err := s.upsertTag(n)
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(
			fmt.Sprintf(`INSERT OR IGNORE INTO %s (%s, tag_id) VALUES (?, ?)`, table, refCol),
			refID, tagID,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) tagsFor(joinTable, refCol, refID string) ([]Tag, error) {
	rows, err := s.db.Query(
		fmt.Sprintf(`SELECT t.id, t.name FROM tags t JOIN %s j ON j.tag_id = t.id WHERE j.%s = ? ORDER BY t.name`, joinTable, refCol),
		refID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// stringSlice coerces a value coming from a JSON map[string]any into []string.
func stringSlice(v any) []string {
	switch xs := v.(type) {
	case []string:
		return xs
	case []any:
		out := make([]string, 0, len(xs))
		for _, x := range xs {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func strField(fields map[string]any, key string) (string, bool) {
	v, ok := fields[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// --- memos ---

func (s *Store) ListMemos() ([]Memo, error) {
	rows, err := s.db.Query(
		`SELECT id, title, content, created_at, updated_at
		 FROM memos WHERE user_id = ? AND deleted_at IS NULL
		 ORDER BY updated_at DESC LIMIT 100`,
		s.userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memos []Memo
	for rows.Next() {
		var m Memo
		if err := rows.Scan(&m.ID, &m.Title, &m.Content, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		memos = append(memos, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range memos {
		tags, err := s.tagsFor("memo_tags", "memo_id", memos[i].ID)
		if err != nil {
			return nil, err
		}
		memos[i].Tags = tags
	}
	return memos, nil
}

func (s *Store) GetMemo(id string) (*Memo, error) {
	var m Memo
	err := s.db.QueryRow(
		`SELECT id, title, content, created_at, updated_at
		 FROM memos WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		id, s.userID,
	).Scan(&m.ID, &m.Title, &m.Content, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	tags, err := s.tagsFor("memo_tags", "memo_id", m.ID)
	if err != nil {
		return nil, err
	}
	m.Tags = tags
	return &m, nil
}

func (s *Store) CreateMemo(title, content string, tags []string) (*Memo, error) {
	now := nowISO()
	id := newID()
	if _, err := s.db.Exec(
		`INSERT INTO memos (id, user_id, title, content, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		id, s.userID, title, content, now, now,
	); err != nil {
		return nil, err
	}
	if len(tags) > 0 {
		if err := s.setTags("memo_tags", "memo_id", id, tags); err != nil {
			return nil, err
		}
	}
	return s.GetMemo(id)
}

func (s *Store) UpdateMemo(id string, fields map[string]any) (*Memo, error) {
	current, err := s.GetMemo(id)
	if err != nil {
		return nil, err
	}
	title := current.Title
	content := current.Content
	if v, ok := strField(fields, "title"); ok {
		title = v
	}
	if v, ok := strField(fields, "content"); ok {
		content = v
	}
	if _, err := s.db.Exec(
		`UPDATE memos SET title = ?, content = ?, updated_at = ?
		 WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		title, content, nowISO(), id, s.userID,
	); err != nil {
		return nil, err
	}
	if raw, ok := fields["tags"]; ok {
		if err := s.setTags("memo_tags", "memo_id", id, stringSlice(raw)); err != nil {
			return nil, err
		}
	}
	return s.GetMemo(id)
}

func (s *Store) DeleteMemo(id string) error {
	_, err := s.db.Exec(
		`UPDATE memos SET deleted_at = ? WHERE id = ? AND user_id = ?`,
		nowISO(), id, s.userID,
	)
	return err
}

// --- todos ---

func (s *Store) ListTodos() ([]Todo, error) {
	rows, err := s.db.Query(
		`SELECT id, title, description, status, COALESCE(due_date, ''), created_at, updated_at
		 FROM todos WHERE user_id = ? AND deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT 100`,
		s.userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range todos {
		tags, err := s.tagsFor("todo_tags", "todo_id", todos[i].ID)
		if err != nil {
			return nil, err
		}
		todos[i].Tags = tags
	}
	return todos, nil
}

func (s *Store) GetTodo(id string) (*Todo, error) {
	var t Todo
	err := s.db.QueryRow(
		`SELECT id, title, description, status, COALESCE(due_date, ''), created_at, updated_at
		 FROM todos WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		id, s.userID,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	tags, err := s.tagsFor("todo_tags", "todo_id", t.ID)
	if err != nil {
		return nil, err
	}
	t.Tags = tags
	return &t, nil
}

func (s *Store) CreateTodo(fields map[string]any) (*Todo, error) {
	title, _ := strField(fields, "title")
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("title is required")
	}
	description, _ := strField(fields, "description")
	status, _ := strField(fields, "status")
	if status == "" {
		status = "open"
	}
	var dueDate sql.NullString
	if v, ok := strField(fields, "due_date"); ok && v != "" {
		dueDate = sql.NullString{String: v, Valid: true}
	}

	now := nowISO()
	id := newID()
	if _, err := s.db.Exec(
		`INSERT INTO todos (id, user_id, title, description, status, due_date, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, s.userID, title, description, status, dueDate, now, now,
	); err != nil {
		return nil, err
	}
	if raw, ok := fields["tags"]; ok {
		if err := s.setTags("todo_tags", "todo_id", id, stringSlice(raw)); err != nil {
			return nil, err
		}
	}
	return s.GetTodo(id)
}

func (s *Store) UpdateTodo(id string, fields map[string]any) (*Todo, error) {
	current, err := s.GetTodo(id)
	if err != nil {
		return nil, err
	}
	title := current.Title
	description := current.Description
	status := current.Status
	dueDate := sql.NullString{String: current.DueDate, Valid: current.DueDate != ""}

	if v, ok := strField(fields, "title"); ok {
		title = v
	}
	if v, ok := strField(fields, "description"); ok {
		description = v
	}
	if v, ok := strField(fields, "status"); ok && v != "" {
		status = v
	}
	if raw, ok := fields["due_date"]; ok {
		if v, ok := raw.(string); ok && v != "" {
			dueDate = sql.NullString{String: v, Valid: true}
		} else {
			dueDate = sql.NullString{}
		}
	}
	if _, err := s.db.Exec(
		`UPDATE todos SET title = ?, description = ?, status = ?, due_date = ?, updated_at = ?
		 WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		title, description, status, dueDate, nowISO(), id, s.userID,
	); err != nil {
		return nil, err
	}
	if raw, ok := fields["tags"]; ok {
		if err := s.setTags("todo_tags", "todo_id", id, stringSlice(raw)); err != nil {
			return nil, err
		}
	}
	return s.GetTodo(id)
}

func (s *Store) SetTodoStatus(id, status string) error {
	res, err := s.db.Exec(
		`UPDATE todos SET status = ?, updated_at = ?
		 WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		status, nowISO(), id, s.userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) DeleteTodo(id string) error {
	_, err := s.db.Exec(
		`UPDATE todos SET deleted_at = ? WHERE id = ? AND user_id = ?`,
		nowISO(), id, s.userID,
	)
	return err
}

// --- events ---

func (s *Store) ListEvents() ([]Event, error) {
	rows, err := s.db.Query(
		`SELECT id, title, description, start_at, COALESCE(end_at, ''), all_day, created_at, updated_at
		 FROM events WHERE user_id = ? AND deleted_at IS NULL
		 ORDER BY start_at ASC LIMIT 100`,
		s.userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var allDay int
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.StartAt, &e.EndAt, &allDay, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		e.AllDay = allDay != 0
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range events {
		tags, err := s.tagsFor("event_tags", "event_id", events[i].ID)
		if err != nil {
			return nil, err
		}
		events[i].Tags = tags
	}
	return events, nil
}

func (s *Store) GetEvent(id string) (*Event, error) {
	var e Event
	var allDay int
	err := s.db.QueryRow(
		`SELECT id, title, description, start_at, COALESCE(end_at, ''), all_day, created_at, updated_at
		 FROM events WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		id, s.userID,
	).Scan(&e.ID, &e.Title, &e.Description, &e.StartAt, &e.EndAt, &allDay, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	e.AllDay = allDay != 0
	tags, err := s.tagsFor("event_tags", "event_id", e.ID)
	if err != nil {
		return nil, err
	}
	e.Tags = tags
	return &e, nil
}

func (s *Store) CreateEvent(fields map[string]any) (*Event, error) {
	title, _ := strField(fields, "title")
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("title is required")
	}
	startAt, _ := strField(fields, "start_at")
	if strings.TrimSpace(startAt) == "" {
		return nil, errors.New("start_at is required")
	}
	description, _ := strField(fields, "description")
	var endAt sql.NullString
	if v, ok := strField(fields, "end_at"); ok && v != "" {
		endAt = sql.NullString{String: v, Valid: true}
	}
	allDay := boolField(fields, "all_day")

	now := nowISO()
	id := newID()
	if _, err := s.db.Exec(
		`INSERT INTO events (id, user_id, title, description, start_at, end_at, all_day, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, s.userID, title, description, startAt, endAt, boolToInt(allDay), now, now,
	); err != nil {
		return nil, err
	}
	if raw, ok := fields["tags"]; ok {
		if err := s.setTags("event_tags", "event_id", id, stringSlice(raw)); err != nil {
			return nil, err
		}
	}
	return s.GetEvent(id)
}

func (s *Store) UpdateEvent(id string, fields map[string]any) (*Event, error) {
	current, err := s.GetEvent(id)
	if err != nil {
		return nil, err
	}
	title := current.Title
	description := current.Description
	startAt := current.StartAt
	endAt := sql.NullString{String: current.EndAt, Valid: current.EndAt != ""}
	allDay := current.AllDay

	if v, ok := strField(fields, "title"); ok {
		title = v
	}
	if v, ok := strField(fields, "description"); ok {
		description = v
	}
	if v, ok := strField(fields, "start_at"); ok && v != "" {
		startAt = v
	}
	if raw, ok := fields["end_at"]; ok {
		if v, ok := raw.(string); ok && v != "" {
			endAt = sql.NullString{String: v, Valid: true}
		} else {
			endAt = sql.NullString{}
		}
	}
	if raw, ok := fields["all_day"]; ok {
		if b, ok := raw.(bool); ok {
			allDay = b
		}
	}
	if _, err := s.db.Exec(
		`UPDATE events SET title = ?, description = ?, start_at = ?, end_at = ?, all_day = ?, updated_at = ?
		 WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		title, description, startAt, endAt, boolToInt(allDay), nowISO(), id, s.userID,
	); err != nil {
		return nil, err
	}
	if raw, ok := fields["tags"]; ok {
		if err := s.setTags("event_tags", "event_id", id, stringSlice(raw)); err != nil {
			return nil, err
		}
	}
	return s.GetEvent(id)
}

func (s *Store) DeleteEvent(id string) error {
	_, err := s.db.Exec(
		`UPDATE events SET deleted_at = ? WHERE id = ? AND user_id = ?`,
		nowISO(), id, s.userID,
	)
	return err
}

func boolField(fields map[string]any, key string) bool {
	v, ok := fields[key]
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// --- search ---

// Search performs a case-insensitive LIKE across memo title/content,
// todo title/description, and event title/description. Postgres' pg_trgm is
// intentionally not reproduced — standalone mode is single-user and the data
// volume stays small.
func (s *Store) Search(query string) ([]SearchResult, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, nil
	}
	like := "%" + q + "%"

	var results []SearchResult

	memoRows, err := s.db.Query(
		`SELECT id, title, content FROM memos
		 WHERE user_id = ? AND deleted_at IS NULL
		   AND (title LIKE ? OR content LIKE ?)
		 ORDER BY updated_at DESC LIMIT 20`,
		s.userID, like, like,
	)
	if err != nil {
		return nil, err
	}
	for memoRows.Next() {
		var id, title, content string
		if err := memoRows.Scan(&id, &title, &content); err != nil {
			memoRows.Close()
			return nil, err
		}
		results = append(results, SearchResult{
			Type: "memo", ID: id, Title: title,
			Snippet: snippet(content, q),
			Tags:    tagNames(s, "memo_tags", "memo_id", id),
		})
	}
	memoRows.Close()

	todoRows, err := s.db.Query(
		`SELECT id, title, description FROM todos
		 WHERE user_id = ? AND deleted_at IS NULL
		   AND (title LIKE ? OR description LIKE ?)
		 ORDER BY created_at DESC LIMIT 20`,
		s.userID, like, like,
	)
	if err != nil {
		return nil, err
	}
	for todoRows.Next() {
		var id, title, desc string
		if err := todoRows.Scan(&id, &title, &desc); err != nil {
			todoRows.Close()
			return nil, err
		}
		results = append(results, SearchResult{
			Type: "todo", ID: id, Title: title,
			Snippet: snippet(desc, q),
			Tags:    tagNames(s, "todo_tags", "todo_id", id),
		})
	}
	todoRows.Close()

	eventRows, err := s.db.Query(
		`SELECT id, title, description FROM events
		 WHERE user_id = ? AND deleted_at IS NULL
		   AND (title LIKE ? OR description LIKE ?)
		 ORDER BY start_at DESC LIMIT 20`,
		s.userID, like, like,
	)
	if err != nil {
		return nil, err
	}
	for eventRows.Next() {
		var id, title, desc string
		if err := eventRows.Scan(&id, &title, &desc); err != nil {
			eventRows.Close()
			return nil, err
		}
		results = append(results, SearchResult{
			Type: "event", ID: id, Title: title,
			Snippet: snippet(desc, q),
			Tags:    tagNames(s, "event_tags", "event_id", id),
		})
	}
	eventRows.Close()

	return results, nil
}

func snippet(text, query string) string {
	if text == "" {
		return ""
	}
	lower := strings.ToLower(text)
	idx := strings.Index(lower, strings.ToLower(query))
	if idx < 0 {
		if len(text) > 120 {
			return text[:120] + "..."
		}
		return text
	}
	start := idx - 40
	if start < 0 {
		start = 0
	}
	end := idx + len(query) + 80
	if end > len(text) {
		end = len(text)
	}
	out := text[start:end]
	if start > 0 {
		out = "..." + out
	}
	if end < len(text) {
		out = out + "..."
	}
	return out
}

func tagNames(s *Store, joinTable, refCol, refID string) []string {
	tags, err := s.tagsFor(joinTable, refCol, refID)
	if err != nil {
		return nil
	}
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return names
}
