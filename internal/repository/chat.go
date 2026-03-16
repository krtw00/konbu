package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ChatSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChatMessage struct {
	ID           uuid.UUID
	SessionID    uuid.UUID
	Role         string
	Content      string
	ToolCalls    json.RawMessage
	ToolCallID   *string
	Provider     *string
	Model        *string
	InputTokens  *int
	OutputTokens *int
	CreatedAt    time.Time
}

func (q *Queries) CreateChatSession(ctx context.Context, userID uuid.UUID, title string) (ChatSession, error) {
	var s ChatSession
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO chat_sessions (user_id, title) VALUES ($1, $2)
		 RETURNING id, user_id, title, created_at, updated_at`,
		userID, title).Scan(&s.ID, &s.UserID, &s.Title, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (q *Queries) ListChatSessionsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]ChatSession, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, user_id, title, created_at, updated_at
		 FROM chat_sessions WHERE user_id = $1 AND deleted_at IS NULL
		 ORDER BY updated_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []ChatSession
	for rows.Next() {
		var s ChatSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.Title, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (q *Queries) GetChatSessionByID(ctx context.Context, id, userID uuid.UUID) (ChatSession, error) {
	var s ChatSession
	err := q.db.QueryRowContext(ctx,
		`SELECT id, user_id, title, created_at, updated_at
		 FROM chat_sessions WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, userID).Scan(&s.ID, &s.UserID, &s.Title, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (q *Queries) UpdateChatSessionTitle(ctx context.Context, id, userID uuid.UUID, title string) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE chat_sessions SET title = $3, updated_at = now()
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, userID, title)
	return err
}

func (q *Queries) TouchChatSession(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE chat_sessions SET updated_at = now() WHERE id = $1`, id)
	return err
}

func (q *Queries) SoftDeleteChatSession(ctx context.Context, id, userID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE chat_sessions SET deleted_at = now() WHERE id = $1 AND user_id = $2`,
		id, userID)
	return err
}

func (q *Queries) CreateChatMessage(ctx context.Context, sessionID uuid.UUID, role, content string, toolCalls json.RawMessage, toolCallID, provider, model *string, inputTokens, outputTokens *int) (ChatMessage, error) {
	var m ChatMessage
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO chat_messages (session_id, role, content, tool_calls, tool_call_id, provider, model, input_tokens, output_tokens)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, session_id, role, content, COALESCE(tool_calls, 'null'::jsonb), tool_call_id, provider, model, input_tokens, output_tokens, created_at`,
		sessionID, role, content, toolCalls, toolCallID, provider, model, inputTokens, outputTokens,
	).Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.ToolCalls, &m.ToolCallID, &m.Provider, &m.Model, &m.InputTokens, &m.OutputTokens, &m.CreatedAt)
	return m, err
}

func (q *Queries) CountUserMessagesThisMonth(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := q.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM chat_messages m
		 JOIN chat_sessions s ON s.id = m.session_id
		 WHERE s.user_id = $1 AND m.role = 'user'
		 AND m.created_at >= date_trunc('month', now())`,
		userID).Scan(&count)
	return count, err
}

func (q *Queries) ListChatMessagesBySessionID(ctx context.Context, sessionID uuid.UUID) ([]ChatMessage, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, session_id, role, content, COALESCE(tool_calls, 'null'::jsonb), tool_call_id, provider, model, input_tokens, output_tokens, created_at
		 FROM chat_messages WHERE session_id = $1
		 ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.ToolCalls, &m.ToolCallID, &m.Provider, &m.Model, &m.InputTokens, &m.OutputTokens, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}
