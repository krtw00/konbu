package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// --- User ---

type User struct {
	ID           uuid.UUID        `json:"id"`
	Email        string           `json:"email"`
	Name         string           `json:"name"`
	IsAdmin      bool             `json:"is_admin"`
	Plan         string           `json:"plan"`
	UserSettings *json.RawMessage `json:"user_settings,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type UpdateSettingsRequest struct {
	Settings json.RawMessage `json:"settings"`
}

// --- API Key ---

type APIKey struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id,omitempty"`
	Name       string     `json:"name"`
	Key        string     `json:"key,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

// --- Tag ---

type Tag struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type TagWithCounts struct {
	Tag
	Counts map[string]int `json:"counts"`
}

type CreateTagRequest struct {
	Name string `json:"name"`
}

type UpdateTagRequest struct {
	Name string `json:"name"`
}

// --- Memo ---

type Memo struct {
	ID           uuid.UUID        `json:"id"`
	UserID       uuid.UUID        `json:"user_id,omitempty"`
	Title        string           `json:"title"`
	Type         string           `json:"type"`
	Content      *string          `json:"content,omitempty"`
	TableColumns *json.RawMessage `json:"table_columns,omitempty"`
	Tags         []Tag            `json:"tags,omitempty"`
	Rows         *PaginatedResult `json:"rows,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type CreateMemoRequest struct {
	Title        string           `json:"title"`
	Type         string           `json:"type"`
	Content      *string          `json:"content,omitempty"`
	TableColumns *json.RawMessage `json:"table_columns,omitempty"`
	Tags         []string         `json:"tags,omitempty"`
}

type UpdateMemoRequest struct {
	Title        string           `json:"title"`
	Content      *string          `json:"content,omitempty"`
	TableColumns *json.RawMessage `json:"table_columns,omitempty"`
	Tags         []string         `json:"tags,omitempty"`
}

// --- Todo ---

type Todo struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	DueDate     *string    `json:"due_date,omitempty"`
	Tags        []Tag      `json:"tags,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateTodoRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	DueDate     *string  `json:"due_date,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type UpdateTodoRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	DueDate     *string  `json:"due_date,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// --- Calendar Event ---

type CalendarEvent struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id,omitempty"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	StartAt        time.Time  `json:"start_at"`
	EndAt          *time.Time `json:"end_at,omitempty"`
	AllDay         bool       `json:"all_day"`
	RecurrenceRule *string    `json:"recurrence_rule,omitempty"`
	RecurrenceEnd  *string    `json:"recurrence_end,omitempty"`
	Tags           []Tag      `json:"tags,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateEventRequest struct {
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	StartAt        time.Time  `json:"start_at"`
	EndAt          *time.Time `json:"end_at,omitempty"`
	AllDay         bool       `json:"all_day"`
	RecurrenceRule *string    `json:"recurrence_rule,omitempty"`
	RecurrenceEnd  *string    `json:"recurrence_end,omitempty"`
	Tags           []string   `json:"tags,omitempty"`
}

type UpdateEventRequest struct {
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	StartAt        time.Time  `json:"start_at"`
	EndAt          *time.Time `json:"end_at,omitempty"`
	AllDay         bool       `json:"all_day"`
	RecurrenceRule *string    `json:"recurrence_rule,omitempty"`
	RecurrenceEnd  *string    `json:"recurrence_end,omitempty"`
	Tags           []string   `json:"tags,omitempty"`
}

// --- Tool ---

type Tool struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id,omitempty"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Icon      string    `json:"icon"`
	Category  string    `json:"category"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateToolRequest struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Category string `json:"category"`
}

type UpdateToolRequest struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Category string `json:"category"`
}

type ReorderRequest struct {
	Order []uuid.UUID `json:"order"`
}

// --- Common ---

type PaginatedResult struct {
	Data   interface{} `json:"data"`
	Total  int64       `json:"total"`
	Limit  int         `json:"limit,omitempty"`
	Offset int         `json:"offset,omitempty"`
}

type ListParams struct {
	Limit  int
	Offset int
	Sort   string
	Query  string
	Tags   []string
}

// --- Search ---

type SearchResult struct {
	Type      string    `json:"type"`
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Snippet   string    `json:"snippet"`
	UpdatedAt time.Time `json:"updated_at"`
}

// --- Chat ---

type ChatSession struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatMessage struct {
	ID           uuid.UUID        `json:"id"`
	Role         string           `json:"role"`
	Content      string           `json:"content"`
	ToolCalls    *json.RawMessage `json:"tool_calls,omitempty"`
	ToolCallID   string           `json:"tool_call_id,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
}

type ChatSessionDetail struct {
	ChatSession
	Messages []ChatMessage `json:"messages"`
}

type SendMessageRequest struct {
	Content string `json:"content"`
}

type UpdateSessionRequest struct {
	Title string `json:"title"`
}

type AIChatConfig struct {
	Provider           string `json:"provider"`
	OpenAIKeyMasked    string `json:"openai_key_masked,omitempty"`
	AnthropicKeyMasked string `json:"anthropic_key_masked,omitempty"`
	OpenAIKey          string `json:"openai_key,omitempty"`
	AnthropicKey       string `json:"anthropic_key,omitempty"`
}

func DefaultListParams() ListParams {
	return ListParams{
		Limit: 20,
		Sort:  "created_at:desc",
	}
}
