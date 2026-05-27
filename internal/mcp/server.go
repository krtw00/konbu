package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func Run(b Backend) error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			writeResponse(os.Stdout, Response{JSONRPC: "2.0", Error: &RPCError{Code: -32700, Message: "parse error"}})
			continue
		}

		// Notifications have no id; per JSON-RPC 2.0, no response should be sent.
		if len(req.ID) == 0 {
			continue
		}

		resp := handle(b, req)
		writeResponse(os.Stdout, resp)
	}

	return scanner.Err()
}

func writeResponse(w io.Writer, resp Response) {
	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "%s\n", data)
}

func handle(b Backend, req Request) Response {
	switch req.Method {
	case "initialize":
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "konbu",
					"version": "1.0.0",
				},
			},
		}

	case "notifications/initialized":
		return Response{} // no response for notifications

	case "tools/list":
		return Response{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{"tools": toolDefinitions()}}

	case "tools/call":
		return handleToolCall(b, req)

	default:
		return Response{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32601, Message: "method not found: " + req.Method}}
	}
}

func handleToolCall(b Backend, req Request) Response {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return Response{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: "invalid params"}}
	}

	result, isErr := executeTool(b, params.Name, params.Arguments)
	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ToolResult{
			Content: []ContentBlock{{Type: "text", Text: result}},
			IsError: isErr,
		},
	}
}

func toJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func executeTool(b Backend, name string, argsRaw json.RawMessage) (string, bool) {
	var args map[string]interface{}
	json.Unmarshal(argsRaw, &args)

	str := func(key string) string {
		v, _ := args[key].(string)
		return v
	}

	switch name {
	case "search":
		results, err := b.Search(str("query"))
		if err != nil {
			return err.Error(), true
		}
		return toJSON(results), false

	case "list_memos":
		memos, err := b.ListMemos()
		if err != nil {
			return err.Error(), true
		}
		return toJSON(memos), false

	case "get_memo":
		memo, err := b.GetMemo(str("id"))
		if err != nil {
			return err.Error(), true
		}
		return toJSON(memo), false

	case "create_memo":
		var tags []string
		if t, ok := args["tags"]; ok {
			if arr, ok := t.([]interface{}); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						tags = append(tags, s)
					}
				}
			}
		}
		memo, err := b.CreateMemo(str("title"), str("content"), tags)
		if err != nil {
			return err.Error(), true
		}
		return toJSON(memo), false

	case "update_memo":
		id := str("id")
		fields := make(map[string]any)
		if v, ok := args["title"]; ok {
			fields["title"] = v
		}
		if v, ok := args["content"]; ok {
			fields["content"] = v
		}
		if v, ok := args["tags"]; ok {
			fields["tags"] = v
		}
		memo, err := b.UpdateMemo(id, fields)
		if err != nil {
			return err.Error(), true
		}
		return toJSON(memo), false

	case "delete_memo":
		if err := b.DeleteMemo(str("id")); err != nil {
			return err.Error(), true
		}
		return "deleted", false

	case "list_todos":
		todos, err := b.ListTodos()
		if err != nil {
			return err.Error(), true
		}
		return toJSON(todos), false

	case "get_todo":
		todo, err := b.GetTodo(str("id"))
		if err != nil {
			return err.Error(), true
		}
		return toJSON(todo), false

	case "create_todo":
		fields := map[string]any{"title": str("title")}
		if v := str("description"); v != "" {
			fields["description"] = v
		}
		if v := str("due_date"); v != "" {
			fields["due_date"] = v
		}
		if v, ok := args["tags"]; ok {
			fields["tags"] = v
		}
		todo, err := b.CreateTodo(fields)
		if err != nil {
			return err.Error(), true
		}
		return toJSON(todo), false

	case "update_todo":
		id := str("id")
		fields := make(map[string]any)
		for _, k := range []string{"title", "description", "status", "due_date"} {
			if v, ok := args[k]; ok {
				fields[k] = v
			}
		}
		if v, ok := args["tags"]; ok {
			fields["tags"] = v
		}
		todo, err := b.UpdateTodo(id, fields)
		if err != nil {
			return err.Error(), true
		}
		return toJSON(todo), false

	case "mark_todo_done":
		if err := b.DoneTodo(str("id")); err != nil {
			return err.Error(), true
		}
		return "marked as done", false

	case "reopen_todo":
		if err := b.ReopenTodo(str("id")); err != nil {
			return err.Error(), true
		}
		return "reopened", false

	case "delete_todo":
		if err := b.DeleteTodo(str("id")); err != nil {
			return err.Error(), true
		}
		return "deleted", false

	case "list_events":
		events, err := b.ListEvents("")
		if err != nil {
			return err.Error(), true
		}
		return toJSON(events), false

	case "get_event":
		event, err := b.GetEvent(str("id"))
		if err != nil {
			return err.Error(), true
		}
		return toJSON(event), false

	case "create_event":
		fields := map[string]any{
			"title":    str("title"),
			"start_at": str("start_at"),
		}
		if v := str("description"); v != "" {
			fields["description"] = v
		}
		if v := str("end_at"); v != "" {
			fields["end_at"] = v
		}
		if v, ok := args["all_day"]; ok {
			fields["all_day"] = v
		}
		if v, ok := args["tags"]; ok {
			fields["tags"] = v
		}
		event, err := b.CreateEvent(fields)
		if err != nil {
			return err.Error(), true
		}
		return toJSON(event), false

	case "update_event":
		id := str("id")
		fields := make(map[string]any)
		for _, k := range []string{"title", "description", "start_at", "end_at"} {
			if v, ok := args[k]; ok {
				fields[k] = v
			}
		}
		if v, ok := args["all_day"]; ok {
			fields["all_day"] = v
		}
		if v, ok := args["tags"]; ok {
			fields["tags"] = v
		}
		event, err := b.UpdateEvent(id, fields)
		if err != nil {
			return err.Error(), true
		}
		return toJSON(event), false

	case "delete_event":
		if err := b.DeleteEvent(str("id")); err != nil {
			return err.Error(), true
		}
		return "deleted", false

	default:
		return fmt.Sprintf("unknown tool: %s", name), true
	}
}

// emptyObjectSchema returns a JSON schema for tools that take no arguments.
func emptyObjectSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func toolDefinitions() []ToolDef {
	return []ToolDef{
		{
			Name: "search",
			Description: "Full-text search across all memos, todos, and calendar events in the user's konbu personal planner. " +
				"Use this when the user asks an open-ended question like \"find anything about X\" or hasn't specified the resource type. " +
				"Returns a unified result list with each match tagged by its resource kind (memo / todo / event). " +
				"For listing every item of a specific kind, prefer list_memos / list_todos / list_events instead.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"query"},
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search keyword. Matches against titles, content/description bodies, and tag names. Whitespace separates terms and is treated as AND.",
					},
				},
			},
		},

		// ----- memos -----
		{
			Name: "list_memos",
			Description: "List all memos (free-form Markdown notes) belonging to the user, newest first. " +
				"Returns id, title, tags, and timestamps for each memo, but NOT the full Markdown body — call get_memo to read a specific memo's full content. " +
				"Memos have no status or due date; for actionable tasks use list_todos instead.",
			InputSchema: emptyObjectSchema(),
		},
		{
			Name:        "get_memo",
			Description: "Fetch the full details of a single memo by ID, including the Markdown content body. Typically used after list_memos or search to read a memo's body.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Memo UUID. Short IDs (first 8 characters of the UUID) are also accepted.",
					},
				},
			},
		},
		{
			Name: "create_memo",
			Description: "Create a new memo (free-form Markdown note) in the konbu planner. " +
				"Returns the created memo with its assigned UUID. " +
				"Use this for notes without a due date or completion state; use create_todo for actionable tasks, and create_event for time-bound calendar items.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"title"},
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Memo title — short, indexed for search. Required.",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Memo body in Markdown. Supports headings, lists, code blocks, and inline tag references such as #project. Optional.",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tag names to attach to the memo. Tags are created on-the-fly if they don't exist yet. Optional.",
					},
				},
			},
		},
		{
			Name: "update_memo",
			Description: "Update an existing memo's title, content, and/or tags. " +
				"Only the fields provided are modified; omitted fields are left unchanged. " +
				"Note that the tags array (if provided) REPLACES the existing tag set — pass the full desired list, not a delta. Returns the updated memo.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the memo to update. Required.",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "New title (optional).",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "New Markdown body (optional). Replaces the existing body entirely.",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "New tag list (optional). Replaces existing tags entirely — pass the full desired list.",
					},
				},
			},
		},
		{
			Name: "delete_memo",
			Description: "Permanently delete a memo by ID. This action cannot be undone. " +
				"Returns the string \"deleted\" on success.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the memo to delete.",
					},
				},
			},
		},

		// ----- todos -----
		{
			Name: "list_todos",
			Description: "List all todos (actionable tasks with status and optional due date), newest first. " +
				"Returns id, title, status (open / done), due_date, and tags for each todo. " +
				"For free-form notes without status use list_memos; for time-bound calendar items use list_events.",
			InputSchema: emptyObjectSchema(),
		},
		{
			Name:        "get_todo",
			Description: "Fetch the full details of a single todo by ID, including description, status, due date, and tags.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Todo UUID. Short IDs (first 8 characters) are also accepted.",
					},
				},
			},
		},
		{
			Name: "create_todo",
			Description: "Create a new todo (actionable task) in the konbu planner. " +
				"Returns the created todo with its assigned UUID and status=\"open\". " +
				"Use this for tasks that need completion tracking; use create_memo for free-form notes, and create_event for items with a specific start time.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"title"},
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Todo title — short imperative phrase, e.g. \"Buy groceries\". Required.",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Longer description or notes for the task (optional).",
					},
					"due_date": map[string]interface{}{
						"type":        "string",
						"description": "Due date in YYYY-MM-DD format, e.g. \"2026-06-15\". Optional.",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tag names to attach (optional). Tags are created on-the-fly if missing.",
					},
				},
			},
		},
		{
			Name: "update_todo",
			Description: "Update an existing todo's fields. Only the fields provided are modified; omitted fields are left unchanged. " +
				"For toggling completion state alone, prefer mark_todo_done / reopen_todo for clearer intent; " +
				"use update_todo when you also need to change title, description, due_date, or tags. " +
				"Note that the tags array (if provided) REPLACES the existing tag set.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the todo to update. Required.",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "New title (optional).",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "New description (optional).",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"open", "done"},
						"description": "Completion status: \"open\" (incomplete) or \"done\" (completed). Prefer mark_todo_done / reopen_todo for status-only changes.",
					},
					"due_date": map[string]interface{}{
						"type":        "string",
						"description": "Due date in YYYY-MM-DD format, e.g. \"2026-06-15\" (optional).",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "New tag list (optional). Replaces existing tags entirely.",
					},
				},
			},
		},
		{
			Name: "mark_todo_done",
			Description: "Mark a todo as completed (status=\"done\"). " +
				"This is a convenience shortcut for the most common state transition; equivalent to update_todo with status=\"done\" but communicates intent more clearly. " +
				"Returns the string \"marked as done\" on success.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the todo to mark as done.",
					},
				},
			},
		},
		{
			Name: "reopen_todo",
			Description: "Reopen a completed todo (transition status from \"done\" back to \"open\"). " +
				"Use this when the user wants to undo a completion. Returns the string \"reopened\" on success.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the completed todo to reopen.",
					},
				},
			},
		},
		{
			Name: "delete_todo",
			Description: "Permanently delete a todo by ID. This action cannot be undone. " +
				"Prefer mark_todo_done if you only want to record completion — completed todos remain searchable and visible in history.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the todo to delete.",
					},
				},
			},
		},

		// ----- events -----
		{
			Name: "list_events",
			Description: "List calendar events belonging to the user, ordered by start time. " +
				"Returns id, title, start_at, end_at, all_day flag, and tags for each event. " +
				"For tasks without a fixed time use list_todos; for free-form notes use list_memos.",
			InputSchema: emptyObjectSchema(),
		},
		{
			Name:        "get_event",
			Description: "Fetch the full details of a single calendar event by ID, including description, all-day flag, and tags.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Event UUID. Short IDs (first 8 characters) are also accepted.",
					},
				},
			},
		},
		{
			Name: "create_event",
			Description: "Create a calendar event with a fixed start time (and optional end time). " +
				"Use this for time-bound items such as meetings or appointments; use create_todo for deadline-style tasks without a specific time, " +
				"and create_memo for free-form notes. Returns the created event with its assigned UUID.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"title", "start_at"},
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Event title. Required.",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Longer description or notes for the event (optional).",
					},
					"start_at": map[string]interface{}{
						"type":        "string",
						"description": "Start datetime in ISO 8601 with timezone offset, e.g. \"2026-06-15T13:00:00+09:00\". Required.",
					},
					"end_at": map[string]interface{}{
						"type":        "string",
						"description": "End datetime in ISO 8601 with timezone offset (optional). Must be after start_at.",
					},
					"all_day": map[string]interface{}{
						"type":        "boolean",
						"description": "If true, the event spans the whole day(s); the time portion of start_at / end_at is ignored.",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tag names to attach (optional).",
					},
				},
			},
		},
		{
			Name: "update_event",
			Description: "Update an existing calendar event. Only the fields provided are modified; omitted fields are left unchanged. " +
				"Note that the tags array (if provided) REPLACES the existing tag set. Returns the updated event.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the event to update. Required.",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "New title (optional).",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "New description (optional).",
					},
					"start_at": map[string]interface{}{
						"type":        "string",
						"description": "New start datetime in ISO 8601 with timezone offset, e.g. \"2026-06-15T13:00:00+09:00\" (optional).",
					},
					"end_at": map[string]interface{}{
						"type":        "string",
						"description": "New end datetime in ISO 8601 with timezone offset (optional). Must be after start_at.",
					},
					"all_day": map[string]interface{}{
						"type":        "boolean",
						"description": "Set the all-day flag (optional).",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "New tag list (optional). Replaces existing tags entirely.",
					},
				},
			},
		},
		{
			Name:        "delete_event",
			Description: "Permanently delete a calendar event by ID. This action cannot be undone.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "UUID of the event to delete.",
					},
				},
			},
		},
	}
}
