package mcp

// Backend is the data layer the MCP server talks to. Two implementations exist:
// the default HTTP backend (talks to a konbu API server) and a future local
// backend that will read/write a SQLite store directly for `konbu mcp --standalone`.
type Backend interface {
	Search(query string) (any, error)

	ListMemos() (any, error)
	GetMemo(id string) (any, error)
	CreateMemo(title, content string, tags []string) (any, error)
	UpdateMemo(id string, fields map[string]any) (any, error)
	DeleteMemo(id string) error

	ListTodos() (any, error)
	GetTodo(id string) (any, error)
	CreateTodo(fields map[string]any) (any, error)
	UpdateTodo(id string, fields map[string]any) (any, error)
	DoneTodo(id string) error
	ReopenTodo(id string) error
	DeleteTodo(id string) error

	ListEvents(calendarID string) (any, error)
	GetEvent(id string) (any, error)
	CreateEvent(fields map[string]any) (any, error)
	UpdateEvent(id string, fields map[string]any) (any, error)
	DeleteEvent(id string) error
}
