package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/krtw00/konbu/internal/client"
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

func Run(cli *client.Client) error {
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

		resp := handle(cli, req)
		writeResponse(os.Stdout, resp)
	}

	return scanner.Err()
}

func writeResponse(w io.Writer, resp Response) {
	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "%s\n", data)
}

func handle(cli *client.Client, req Request) Response {
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
		return handleToolCall(cli, req)

	default:
		return Response{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32601, Message: "method not found: " + req.Method}}
	}
}

func handleToolCall(cli *client.Client, req Request) Response {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return Response{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: "invalid params"}}
	}

	result, isErr := executeTool(cli, params.Name, params.Arguments)
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

func executeTool(cli *client.Client, name string, argsRaw json.RawMessage) (string, bool) {
	var args map[string]interface{}
	json.Unmarshal(argsRaw, &args)

	str := func(key string) string {
		v, _ := args[key].(string)
		return v
	}

	switch name {
	case "search":
		results, err := cli.Search(str("query"))
		if err != nil {
			return err.Error(), true
		}
		return toJSON(results), false

	case "list_memos":
		memos, err := cli.ListMemos()
		if err != nil {
			return err.Error(), true
		}
		return toJSON(memos), false

	case "get_memo":
		memo, err := cli.GetMemo(str("id"))
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
		memo, err := cli.CreateMemo(str("title"), str("content"), tags)
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
		memo, err := cli.UpdateMemo(id, fields)
		if err != nil {
			return err.Error(), true
		}
		return toJSON(memo), false

	case "delete_memo":
		if err := cli.DeleteMemo(str("id")); err != nil {
			return err.Error(), true
		}
		return "deleted", false

	case "list_todos":
		todos, err := cli.ListTodos()
		if err != nil {
			return err.Error(), true
		}
		return toJSON(todos), false

	case "get_todo":
		todo, err := cli.GetTodo(str("id"))
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
		todo, err := cli.CreateTodo(fields)
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
		todo, err := cli.UpdateTodo(id, fields)
		if err != nil {
			return err.Error(), true
		}
		return toJSON(todo), false

	case "mark_todo_done":
		if err := cli.DoneTodo(str("id")); err != nil {
			return err.Error(), true
		}
		return "marked as done", false

	case "reopen_todo":
		if err := cli.ReopenTodo(str("id")); err != nil {
			return err.Error(), true
		}
		return "reopened", false

	case "delete_todo":
		if err := cli.DeleteTodo(str("id")); err != nil {
			return err.Error(), true
		}
		return "deleted", false

	case "list_events":
		events, err := cli.ListEvents("")
		if err != nil {
			return err.Error(), true
		}
		return toJSON(events), false

	case "get_event":
		event, err := cli.GetEvent(str("id"))
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
		event, err := cli.CreateEvent(fields)
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
		event, err := cli.UpdateEvent(id, fields)
		if err != nil {
			return err.Error(), true
		}
		return toJSON(event), false

	case "delete_event":
		if err := cli.DeleteEvent(str("id")); err != nil {
			return err.Error(), true
		}
		return "deleted", false

	default:
		return fmt.Sprintf("unknown tool: %s", name), true
	}
}

func s(desc string) map[string]interface{} {
	return map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "description": desc}
}

func toolDefinitions() []ToolDef {
	return []ToolDef{
		{Name: "search", Description: "メモ・ToDo・イベントを横断検索", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"query"},
			"properties": map[string]interface{}{"query": map[string]interface{}{"type": "string", "description": "検索キーワード"}},
		}},
		{Name: "list_memos", Description: "メモ一覧を取得", InputSchema: s("メモ一覧")},
		{Name: "get_memo", Description: "メモの詳細を取得", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"id"},
			"properties": map[string]interface{}{"id": map[string]interface{}{"type": "string", "description": "メモID"}},
		}},
		{Name: "create_memo", Description: "新しいメモを作成", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"title"},
			"properties": map[string]interface{}{
				"title":   map[string]interface{}{"type": "string"},
				"content": map[string]interface{}{"type": "string"},
				"tags":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			},
		}},
		{Name: "update_memo", Description: "メモを更新", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"id"},
			"properties": map[string]interface{}{
				"id":      map[string]interface{}{"type": "string"},
				"title":   map[string]interface{}{"type": "string"},
				"content": map[string]interface{}{"type": "string"},
				"tags":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			},
		}},
		{Name: "delete_memo", Description: "メモを削除", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"id"},
			"properties": map[string]interface{}{"id": map[string]interface{}{"type": "string"}},
		}},
		{Name: "list_todos", Description: "ToDo一覧を取得", InputSchema: s("ToDo一覧")},
		{Name: "get_todo", Description: "ToDoの詳細を取得", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"id"},
			"properties": map[string]interface{}{"id": map[string]interface{}{"type": "string"}},
		}},
		{Name: "create_todo", Description: "新しいToDoを作成", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"title"},
			"properties": map[string]interface{}{
				"title":       map[string]interface{}{"type": "string"},
				"description": map[string]interface{}{"type": "string"},
				"due_date":    map[string]interface{}{"type": "string", "description": "YYYY-MM-DD"},
				"tags":        map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			},
		}},
		{Name: "update_todo", Description: "ToDoを更新", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"id"},
			"properties": map[string]interface{}{
				"id":          map[string]interface{}{"type": "string"},
				"title":       map[string]interface{}{"type": "string"},
				"description": map[string]interface{}{"type": "string"},
				"status":      map[string]interface{}{"type": "string", "enum": []string{"open", "done"}},
				"due_date":    map[string]interface{}{"type": "string"},
				"tags":        map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			},
		}},
		{Name: "mark_todo_done", Description: "ToDoを完了にする", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"id"},
			"properties": map[string]interface{}{"id": map[string]interface{}{"type": "string"}},
		}},
		{Name: "reopen_todo", Description: "完了したToDoを再開", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"id"},
			"properties": map[string]interface{}{"id": map[string]interface{}{"type": "string"}},
		}},
		{Name: "delete_todo", Description: "ToDoを削除", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"id"},
			"properties": map[string]interface{}{"id": map[string]interface{}{"type": "string"}},
		}},
		{Name: "list_events", Description: "カレンダーイベント一覧を取得", InputSchema: s("イベント一覧")},
		{Name: "get_event", Description: "イベントの詳細を取得", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"id"},
			"properties": map[string]interface{}{"id": map[string]interface{}{"type": "string"}},
		}},
		{Name: "create_event", Description: "新しいイベントを作成", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"title", "start_at"},
			"properties": map[string]interface{}{
				"title":       map[string]interface{}{"type": "string"},
				"description": map[string]interface{}{"type": "string"},
				"start_at":    map[string]interface{}{"type": "string", "description": "ISO 8601"},
				"end_at":      map[string]interface{}{"type": "string", "description": "ISO 8601"},
				"all_day":     map[string]interface{}{"type": "boolean"},
				"tags":        map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			},
		}},
		{Name: "update_event", Description: "イベントを更新", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"id"},
			"properties": map[string]interface{}{
				"id":          map[string]interface{}{"type": "string"},
				"title":       map[string]interface{}{"type": "string"},
				"description": map[string]interface{}{"type": "string"},
				"start_at":    map[string]interface{}{"type": "string"},
				"end_at":      map[string]interface{}{"type": "string"},
				"all_day":     map[string]interface{}{"type": "boolean"},
				"tags":        map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			},
		}},
		{Name: "delete_event", Description: "イベントを削除", InputSchema: map[string]interface{}{
			"type": "object", "required": []string{"id"},
			"properties": map[string]interface{}{"id": map[string]interface{}{"type": "string"}},
		}},
	}
}
