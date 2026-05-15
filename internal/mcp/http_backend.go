package mcp

import "github.com/krtw00/konbu/internal/client"

type httpBackend struct {
	cli *client.Client
}

// NewHTTPBackend wraps a konbu API HTTP client as a Backend.
func NewHTTPBackend(cli *client.Client) Backend {
	return &httpBackend{cli: cli}
}

func (b *httpBackend) Search(query string) (any, error) { return b.cli.Search(query) }

func (b *httpBackend) ListMemos() (any, error)         { return b.cli.ListMemos() }
func (b *httpBackend) GetMemo(id string) (any, error)  { return b.cli.GetMemo(id) }
func (b *httpBackend) CreateMemo(title, content string, tags []string) (any, error) {
	return b.cli.CreateMemo(title, content, tags)
}
func (b *httpBackend) UpdateMemo(id string, fields map[string]any) (any, error) {
	return b.cli.UpdateMemo(id, fields)
}
func (b *httpBackend) DeleteMemo(id string) error { return b.cli.DeleteMemo(id) }

func (b *httpBackend) ListTodos() (any, error)        { return b.cli.ListTodos() }
func (b *httpBackend) GetTodo(id string) (any, error) { return b.cli.GetTodo(id) }
func (b *httpBackend) CreateTodo(fields map[string]any) (any, error) {
	return b.cli.CreateTodo(fields)
}
func (b *httpBackend) UpdateTodo(id string, fields map[string]any) (any, error) {
	return b.cli.UpdateTodo(id, fields)
}
func (b *httpBackend) DoneTodo(id string) error   { return b.cli.DoneTodo(id) }
func (b *httpBackend) ReopenTodo(id string) error { return b.cli.ReopenTodo(id) }
func (b *httpBackend) DeleteTodo(id string) error { return b.cli.DeleteTodo(id) }

func (b *httpBackend) ListEvents(calendarID string) (any, error) {
	return b.cli.ListEvents(calendarID)
}
func (b *httpBackend) GetEvent(id string) (any, error) { return b.cli.GetEvent(id) }
func (b *httpBackend) CreateEvent(fields map[string]any) (any, error) {
	return b.cli.CreateEvent(fields)
}
func (b *httpBackend) UpdateEvent(id string, fields map[string]any) (any, error) {
	return b.cli.UpdateEvent(id, fields)
}
func (b *httpBackend) DeleteEvent(id string) error { return b.cli.DeleteEvent(id) }
