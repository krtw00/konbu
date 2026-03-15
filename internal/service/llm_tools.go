package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/model"
)

// jsonSchema is a helper for building JSON Schema objects inline.
type jsonSchema map[string]interface{}

var KonbuTools = []ToolDefinition{
	{
		Name:        "search",
		Description: "メモ・ToDo・イベントを横断検索する",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"query": jsonSchema{"type": "string", "description": "検索キーワード（2文字以上）"},
				"limit": jsonSchema{"type": "integer", "description": "最大件数（デフォルト20、上限50）"},
			},
			"required": []string{"query"},
		},
	},
	{
		Name:        "list_memos",
		Description: "メモ一覧を取得する（絞り込み可能）",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"limit":  jsonSchema{"type": "integer", "description": "取得件数"},
				"offset": jsonSchema{"type": "integer", "description": "オフセット"},
				"q":      jsonSchema{"type": "string", "description": "検索キーワード"},
				"tag":    jsonSchema{"type": "string", "description": "タグで絞り込み"},
			},
		},
	},
	{
		Name:        "create_memo",
		Description: "メモを新規作成する",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"title":   jsonSchema{"type": "string", "description": "メモのタイトル"},
				"content": jsonSchema{"type": "string", "description": "メモの本文（Markdown）"},
				"tags":    jsonSchema{"type": "array", "items": jsonSchema{"type": "string"}, "description": "タグ名の配列"},
			},
			"required": []string{"title"},
		},
	},
	{
		Name:        "update_memo",
		Description: "既存のメモを更新する",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"id":      jsonSchema{"type": "string", "format": "uuid", "description": "メモのID"},
				"title":   jsonSchema{"type": "string", "description": "タイトル"},
				"content": jsonSchema{"type": "string", "description": "本文（Markdown）"},
				"tags":    jsonSchema{"type": "array", "items": jsonSchema{"type": "string"}, "description": "タグ名の配列"},
			},
			"required": []string{"id"},
		},
	},
	{
		Name:        "delete_memo",
		Description: "メモを削除する",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"id": jsonSchema{"type": "string", "format": "uuid", "description": "メモのID"},
			},
			"required": []string{"id"},
		},
	},
	{
		Name:        "list_todos",
		Description: "ToDo一覧を取得する（絞り込み可能）",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"limit":  jsonSchema{"type": "integer", "description": "取得件数"},
				"offset": jsonSchema{"type": "integer", "description": "オフセット"},
				"q":      jsonSchema{"type": "string", "description": "検索キーワード"},
				"tag":    jsonSchema{"type": "string", "description": "タグで絞り込み"},
			},
		},
	},
	{
		Name:        "create_todo",
		Description: "ToDoを新規作成する",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"title":       jsonSchema{"type": "string", "description": "ToDoのタイトル"},
				"description": jsonSchema{"type": "string", "description": "詳細説明"},
				"due_date":    jsonSchema{"type": "string", "format": "date", "description": "期限（YYYY-MM-DD）"},
				"tags":        jsonSchema{"type": "array", "items": jsonSchema{"type": "string"}, "description": "タグ名の配列"},
			},
			"required": []string{"title"},
		},
	},
	{
		Name:        "update_todo",
		Description: "既存のToDoを更新する",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"id":          jsonSchema{"type": "string", "format": "uuid", "description": "ToDoのID"},
				"title":       jsonSchema{"type": "string", "description": "タイトル"},
				"description": jsonSchema{"type": "string", "description": "詳細説明"},
				"status":      jsonSchema{"type": "string", "enum": []string{"open", "done"}, "description": "ステータス"},
				"due_date":    jsonSchema{"type": "string", "format": "date", "description": "期限（YYYY-MM-DD）"},
				"tags":        jsonSchema{"type": "array", "items": jsonSchema{"type": "string"}, "description": "タグ名の配列"},
			},
			"required": []string{"id"},
		},
	},
	{
		Name:        "mark_todo_done",
		Description: "ToDoを完了にする",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"id": jsonSchema{"type": "string", "format": "uuid", "description": "ToDoのID"},
			},
			"required": []string{"id"},
		},
	},
	{
		Name:        "reopen_todo",
		Description: "完了済みのToDoを再開する",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"id": jsonSchema{"type": "string", "format": "uuid", "description": "ToDoのID"},
			},
			"required": []string{"id"},
		},
	},
	{
		Name:        "delete_todo",
		Description: "ToDoを削除する",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"id": jsonSchema{"type": "string", "format": "uuid", "description": "ToDoのID"},
			},
			"required": []string{"id"},
		},
	},
	{
		Name:        "list_events",
		Description: "イベント一覧を取得する（絞り込み可能）",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"limit":  jsonSchema{"type": "integer", "description": "取得件数"},
				"offset": jsonSchema{"type": "integer", "description": "オフセット"},
				"q":      jsonSchema{"type": "string", "description": "検索キーワード"},
				"tag":    jsonSchema{"type": "string", "description": "タグで絞り込み"},
			},
		},
	},
	{
		Name:        "create_event",
		Description: "イベントを新規作成する",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"title":           jsonSchema{"type": "string", "description": "イベントのタイトル"},
				"description":     jsonSchema{"type": "string", "description": "詳細説明"},
				"start_at":        jsonSchema{"type": "string", "format": "date-time", "description": "開始日時（RFC3339）"},
				"end_at":          jsonSchema{"type": "string", "format": "date-time", "description": "終了日時（RFC3339）"},
				"all_day":         jsonSchema{"type": "boolean", "description": "終日イベントかどうか"},
				"recurrence_rule": jsonSchema{"type": "string", "description": "繰り返しルール（RRULE形式）"},
				"recurrence_end":  jsonSchema{"type": "string", "format": "date", "description": "繰り返し終了日（YYYY-MM-DD）"},
				"tags":            jsonSchema{"type": "array", "items": jsonSchema{"type": "string"}, "description": "タグ名の配列"},
			},
			"required": []string{"title", "start_at"},
		},
	},
	{
		Name:        "update_event",
		Description: "既存のイベントを更新する",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"id":              jsonSchema{"type": "string", "format": "uuid", "description": "イベントのID"},
				"title":           jsonSchema{"type": "string", "description": "タイトル"},
				"description":     jsonSchema{"type": "string", "description": "詳細説明"},
				"start_at":        jsonSchema{"type": "string", "format": "date-time", "description": "開始日時（RFC3339）"},
				"end_at":          jsonSchema{"type": "string", "format": "date-time", "description": "終了日時（RFC3339）"},
				"all_day":         jsonSchema{"type": "boolean", "description": "終日イベントかどうか"},
				"recurrence_rule": jsonSchema{"type": "string", "description": "繰り返しルール（RRULE形式）"},
				"recurrence_end":  jsonSchema{"type": "string", "format": "date", "description": "繰り返し終了日（YYYY-MM-DD）"},
				"tags":            jsonSchema{"type": "array", "items": jsonSchema{"type": "string"}, "description": "タグ名の配列"},
			},
			"required": []string{"id", "title", "start_at"},
		},
	},
	{
		Name:        "delete_event",
		Description: "イベントを削除する",
		Parameters: jsonSchema{
			"type": "object",
			"properties": jsonSchema{
				"id": jsonSchema{"type": "string", "format": "uuid", "description": "イベントのID"},
			},
			"required": []string{"id"},
		},
	},
}

type ToolExecutor struct {
	memoSvc   *MemoService
	todoSvc   *TodoService
	eventSvc  *EventService
	searchSvc *SearchService
}

func NewToolExecutor(memoSvc *MemoService, todoSvc *TodoService, eventSvc *EventService, searchSvc *SearchService) *ToolExecutor {
	return &ToolExecutor{
		memoSvc:   memoSvc,
		todoSvc:   todoSvc,
		eventSvc:  eventSvc,
		searchSvc: searchSvc,
	}
}

func (e *ToolExecutor) Execute(ctx context.Context, userID uuid.UUID, toolName, argsJSON string) (string, error) {
	var args map[string]interface{}
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return toolError("invalid JSON arguments: " + err.Error()), nil
		}
	}
	if args == nil {
		args = map[string]interface{}{}
	}

	switch toolName {
	case "search":
		return e.execSearch(ctx, userID, args)
	case "list_memos":
		return e.execListMemos(ctx, userID, args)
	case "create_memo":
		return e.execCreateMemo(ctx, userID, args)
	case "update_memo":
		return e.execUpdateMemo(ctx, userID, args)
	case "delete_memo":
		return e.execDeleteMemo(ctx, userID, args)
	case "list_todos":
		return e.execListTodos(ctx, userID, args)
	case "create_todo":
		return e.execCreateTodo(ctx, userID, args)
	case "update_todo":
		return e.execUpdateTodo(ctx, userID, args)
	case "mark_todo_done":
		return e.execMarkTodoDone(ctx, userID, args)
	case "reopen_todo":
		return e.execReopenTodo(ctx, userID, args)
	case "delete_todo":
		return e.execDeleteTodo(ctx, userID, args)
	case "list_events":
		return e.execListEvents(ctx, userID, args)
	case "create_event":
		return e.execCreateEvent(ctx, userID, args)
	case "update_event":
		return e.execUpdateEvent(ctx, userID, args)
	case "delete_event":
		return e.execDeleteEvent(ctx, userID, args)
	default:
		return toolError(fmt.Sprintf("unknown tool: %s", toolName)), nil
	}
}

// --- search ---

func (e *ToolExecutor) execSearch(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return toolError("query is required"), nil
	}
	limit := intArg(args, "limit", 20)
	results, err := e.searchSvc.Search(ctx, userID, query, limit)
	if err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(results)
}

// --- memo ---

func (e *ToolExecutor) execListMemos(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	params := buildListParams(args)
	result, err := e.memoSvc.ListMemos(ctx, userID, params)
	if err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(result)
}

func (e *ToolExecutor) execCreateMemo(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	title, _ := args["title"].(string)
	if title == "" {
		return toolError("title is required"), nil
	}
	req := model.CreateMemoRequest{
		Title: title,
		Type:  "markdown",
		Tags:  stringSliceArg(args, "tags"),
	}
	if c, ok := args["content"].(string); ok {
		req.Content = &c
	}
	memo, err := e.memoSvc.CreateMemo(ctx, userID, req)
	if err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(memo)
}

func (e *ToolExecutor) execUpdateMemo(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	id, err := uuidArg(args, "id")
	if err != nil {
		return toolError("id is required and must be a valid UUID"), nil
	}
	req := model.UpdateMemoRequest{
		Tags: stringSliceArg(args, "tags"),
	}
	if t, ok := args["title"].(string); ok {
		req.Title = t
	}
	if c, ok := args["content"].(string); ok {
		req.Content = &c
	}
	memo, err := e.memoSvc.UpdateMemo(ctx, id, userID, req)
	if err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(memo)
}

func (e *ToolExecutor) execDeleteMemo(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	id, err := uuidArg(args, "id")
	if err != nil {
		return toolError("id is required and must be a valid UUID"), nil
	}
	if err := e.memoSvc.DeleteMemo(ctx, id, userID); err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(map[string]string{"status": "deleted"})
}

// --- todo ---

func (e *ToolExecutor) execListTodos(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	params := buildListParams(args)
	result, err := e.todoSvc.ListTodos(ctx, userID, params)
	if err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(result)
}

func (e *ToolExecutor) execCreateTodo(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	title, _ := args["title"].(string)
	if title == "" {
		return toolError("title is required"), nil
	}
	req := model.CreateTodoRequest{
		Title:       title,
		Description: stringArg(args, "description"),
		Tags:        stringSliceArg(args, "tags"),
	}
	if d, ok := args["due_date"].(string); ok && d != "" {
		req.DueDate = &d
	}
	todo, err := e.todoSvc.CreateTodo(ctx, userID, req)
	if err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(todo)
}

func (e *ToolExecutor) execUpdateTodo(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	id, err := uuidArg(args, "id")
	if err != nil {
		return toolError("id is required and must be a valid UUID"), nil
	}
	req := model.UpdateTodoRequest{
		Title:       stringArg(args, "title"),
		Description: stringArg(args, "description"),
		Status:      stringArg(args, "status"),
		Tags:        stringSliceArg(args, "tags"),
	}
	if d, ok := args["due_date"].(string); ok && d != "" {
		req.DueDate = &d
	}
	todo, err := e.todoSvc.UpdateTodo(ctx, id, userID, req)
	if err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(todo)
}

func (e *ToolExecutor) execMarkTodoDone(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	id, err := uuidArg(args, "id")
	if err != nil {
		return toolError("id is required and must be a valid UUID"), nil
	}
	if err := e.todoSvc.MarkDone(ctx, id, userID); err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(map[string]string{"status": "done"})
}

func (e *ToolExecutor) execReopenTodo(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	id, err := uuidArg(args, "id")
	if err != nil {
		return toolError("id is required and must be a valid UUID"), nil
	}
	if err := e.todoSvc.Reopen(ctx, id, userID); err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(map[string]string{"status": "reopened"})
}

func (e *ToolExecutor) execDeleteTodo(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	id, err := uuidArg(args, "id")
	if err != nil {
		return toolError("id is required and must be a valid UUID"), nil
	}
	if err := e.todoSvc.DeleteTodo(ctx, id, userID); err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(map[string]string{"status": "deleted"})
}

// --- event ---

func (e *ToolExecutor) execListEvents(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	params := buildListParams(args)
	result, err := e.eventSvc.ListEvents(ctx, userID, params)
	if err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(result)
}

func (e *ToolExecutor) execCreateEvent(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	title, _ := args["title"].(string)
	if title == "" {
		return toolError("title is required"), nil
	}
	startAt, err := timeArg(args, "start_at")
	if err != nil {
		return toolError("start_at is required (RFC3339 format)"), nil
	}
	req := model.CreateEventRequest{
		Title:       title,
		Description: stringArg(args, "description"),
		StartAt:     startAt,
		AllDay:      boolArg(args, "all_day"),
		Tags:        stringSliceArg(args, "tags"),
	}
	if v, ok := args["end_at"].(string); ok && v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return toolError("end_at must be RFC3339 format"), nil
		}
		req.EndAt = &t
	}
	if v, ok := args["recurrence_rule"].(string); ok && v != "" {
		req.RecurrenceRule = &v
	}
	if v, ok := args["recurrence_end"].(string); ok && v != "" {
		req.RecurrenceEnd = &v
	}
	event, err := e.eventSvc.CreateEvent(ctx, userID, req)
	if err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(event)
}

func (e *ToolExecutor) execUpdateEvent(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	id, err := uuidArg(args, "id")
	if err != nil {
		return toolError("id is required and must be a valid UUID"), nil
	}
	title, _ := args["title"].(string)
	startAt, err := timeArg(args, "start_at")
	if err != nil {
		return toolError("start_at is required (RFC3339 format)"), nil
	}
	req := model.UpdateEventRequest{
		Title:       title,
		Description: stringArg(args, "description"),
		StartAt:     startAt,
		AllDay:      boolArg(args, "all_day"),
		Tags:        stringSliceArg(args, "tags"),
	}
	if v, ok := args["end_at"].(string); ok && v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return toolError("end_at must be RFC3339 format"), nil
		}
		req.EndAt = &t
	}
	if v, ok := args["recurrence_rule"].(string); ok && v != "" {
		req.RecurrenceRule = &v
	}
	if v, ok := args["recurrence_end"].(string); ok && v != "" {
		req.RecurrenceEnd = &v
	}
	event, err := e.eventSvc.UpdateEvent(ctx, id, userID, req)
	if err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(event)
}

func (e *ToolExecutor) execDeleteEvent(ctx context.Context, userID uuid.UUID, args map[string]interface{}) (string, error) {
	id, err := uuidArg(args, "id")
	if err != nil {
		return toolError("id is required and must be a valid UUID"), nil
	}
	if err := e.eventSvc.DeleteEvent(ctx, id, userID); err != nil {
		return toolError(err.Error()), nil
	}
	return toJSON(map[string]string{"status": "deleted"})
}

// --- helpers ---

func buildListParams(args map[string]interface{}) model.ListParams {
	params := model.DefaultListParams()
	if v := intArg(args, "limit", 0); v > 0 {
		params.Limit = v
	}
	if v := intArg(args, "offset", 0); v > 0 {
		params.Offset = v
	}
	if v, ok := args["q"].(string); ok && v != "" {
		params.Query = v
	}
	if v, ok := args["tag"].(string); ok && v != "" {
		params.Tags = []string{v}
	}
	return params
}

func intArg(args map[string]interface{}, key string, defaultVal int) int {
	v, ok := args[key]
	if !ok {
		return defaultVal
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return defaultVal
	}
}

func stringArg(args map[string]interface{}, key string) string {
	v, _ := args[key].(string)
	return v
}

func boolArg(args map[string]interface{}, key string) bool {
	v, _ := args[key].(bool)
	return v
}

func stringSliceArg(args map[string]interface{}, key string) []string {
	v, ok := args[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func uuidArg(args map[string]interface{}, key string) (uuid.UUID, error) {
	v, ok := args[key].(string)
	if !ok || v == "" {
		return uuid.Nil, fmt.Errorf("%s is required", key)
	}
	return uuid.Parse(v)
}

func timeArg(args map[string]interface{}, key string) (time.Time, error) {
	v, ok := args[key].(string)
	if !ok || v == "" {
		return time.Time{}, fmt.Errorf("%s is required", key)
	}
	return time.Parse(time.RFC3339, v)
}

func toJSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return toolError("failed to marshal result: " + err.Error()), nil
	}
	return string(b), nil
}

func toolError(msg string) string {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return string(b)
}
