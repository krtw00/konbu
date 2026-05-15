package mcp

import "github.com/krtw00/konbu/internal/standalone"

type localBackend struct {
	store *standalone.Store
}

// NewLocalBackend wraps a standalone SQLite store as a Backend, for use by
// `konbu mcp --standalone`. The caller owns the store and is responsible for
// closing it.
func NewLocalBackend(store *standalone.Store) Backend {
	return &localBackend{store: store}
}

func (b *localBackend) Search(query string) (any, error) { return b.store.Search(query) }

func (b *localBackend) ListMemos() (any, error)        { return b.store.ListMemos() }
func (b *localBackend) GetMemo(id string) (any, error) { return b.store.GetMemo(id) }
func (b *localBackend) CreateMemo(title, content string, tags []string) (any, error) {
	return b.store.CreateMemo(title, content, tags)
}
func (b *localBackend) UpdateMemo(id string, fields map[string]any) (any, error) {
	return b.store.UpdateMemo(id, fields)
}
func (b *localBackend) DeleteMemo(id string) error { return b.store.DeleteMemo(id) }

func (b *localBackend) ListTodos() (any, error)        { return b.store.ListTodos() }
func (b *localBackend) GetTodo(id string) (any, error) { return b.store.GetTodo(id) }
func (b *localBackend) CreateTodo(fields map[string]any) (any, error) {
	return b.store.CreateTodo(fields)
}
func (b *localBackend) UpdateTodo(id string, fields map[string]any) (any, error) {
	return b.store.UpdateTodo(id, fields)
}
func (b *localBackend) DoneTodo(id string) error   { return b.store.SetTodoStatus(id, "done") }
func (b *localBackend) ReopenTodo(id string) error { return b.store.SetTodoStatus(id, "open") }
func (b *localBackend) DeleteTodo(id string) error { return b.store.DeleteTodo(id) }

// ListEvents ignores calendarID — standalone mode has no calendars resource.
func (b *localBackend) ListEvents(calendarID string) (any, error) { return b.store.ListEvents() }
func (b *localBackend) GetEvent(id string) (any, error)           { return b.store.GetEvent(id) }
func (b *localBackend) CreateEvent(fields map[string]any) (any, error) {
	return b.store.CreateEvent(fields)
}
func (b *localBackend) UpdateEvent(id string, fields map[string]any) (any, error) {
	return b.store.UpdateEvent(id, fields)
}
func (b *localBackend) DeleteEvent(id string) error { return b.store.DeleteEvent(id) }
