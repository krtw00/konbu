package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/config"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

const maxToolCalls = 5

type ChatService struct {
	queries  *repository.Queries
	db       *sql.DB
	cfg      *config.Config
	executor *ToolExecutor
}

func NewChatService(db *sql.DB, cfg *config.Config, memoSvc *MemoService, todoSvc *TodoService, eventSvc *EventService, searchSvc *SearchService) *ChatService {
	return &ChatService{
		queries:  repository.New(db),
		db:       db,
		cfg:      cfg,
		executor: NewToolExecutor(memoSvc, todoSvc, eventSvc, searchSvc),
	}
}

func (s *ChatService) ListSessions(ctx context.Context, userID uuid.UUID) ([]model.ChatSession, error) {
	rows, err := s.queries.ListChatSessionsByUserID(ctx, userID, 50, 0)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	sessions := make([]model.ChatSession, len(rows))
	for i, r := range rows {
		sessions[i] = model.ChatSession{ID: r.ID, Title: r.Title, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
	}
	return sessions, nil
}

func (s *ChatService) GetSession(ctx context.Context, sessionID, userID uuid.UUID) (*model.ChatSessionDetail, error) {
	sess, err := s.queries.GetChatSessionByID(ctx, sessionID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("chat session")
		}
		return nil, apperror.Internal(err)
	}

	rows, err := s.queries.ListChatMessagesBySessionIDForUser(ctx, sessionID, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	msgs := make([]model.ChatMessage, len(rows))
	for i, r := range rows {
		msgs[i] = toModelChatMessage(r)
	}

	return &model.ChatSessionDetail{
		ChatSession: model.ChatSession{ID: sess.ID, Title: sess.Title, CreatedAt: sess.CreatedAt, UpdatedAt: sess.UpdatedAt},
		Messages:    msgs,
	}, nil
}

func (s *ChatService) CreateSession(ctx context.Context, userID uuid.UUID) (*model.ChatSession, error) {
	sess, err := s.queries.CreateChatSession(ctx, userID, "")
	if err != nil {
		return nil, apperror.Internal(err)
	}
	result := model.ChatSession{ID: sess.ID, Title: sess.Title, CreatedAt: sess.CreatedAt, UpdatedAt: sess.UpdatedAt}
	return &result, nil
}

func (s *ChatService) UpdateSessionTitle(ctx context.Context, sessionID, userID uuid.UUID, title string) error {
	return s.queries.UpdateChatSessionTitle(ctx, sessionID, userID, title)
}

func (s *ChatService) DeleteSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	return s.queries.SoftDeleteChatSession(ctx, sessionID, userID)
}

type SSEEvent struct {
	Event string
	Data  interface{}
}

func (s *ChatService) SendMessage(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, content string, user *model.User) (<-chan SSEEvent, error) {
	sess, err := s.queries.GetChatSessionByID(ctx, sessionID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("chat session")
		}
		return nil, apperror.Internal(err)
	}

	// Get AI config
	provider, apiKey, err := s.getAIKey(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Rate limit when using server default key
	usingDefaultKey := s.cfg.DefaultAIAPIKey != "" && apiKey == s.cfg.DefaultAIAPIKey
	if usingDefaultKey && !user.IsAdmin {
		count, err := s.queries.CountUserMessagesThisMonth(ctx, userID)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		limit := 20
		if user.Plan == "sponsor" {
			limit = 100
		}
		if count >= limit {
			return nil, apperror.Forbidden(fmt.Sprintf("Monthly AI chat limit reached (%d/%d). Set your own API key in Settings for unlimited access.", count, limit))
		}
	}

	// Save user message
	if _, err := s.queries.CreateChatMessage(ctx, sessionID, "user", content, nil, nil, nil, nil, nil, nil); err != nil {
		return nil, apperror.Internal(err)
	}
	if err := s.queries.TouchChatSessionByUser(ctx, sessionID, userID); err != nil {
		return nil, apperror.Internal(err)
	}

	// Update session title if empty
	if sess.Title == "" {
		title := content
		if len([]rune(title)) > 30 {
			title = string([]rune(title)[:30]) + "..."
		}
		if err := s.queries.UpdateChatSessionTitle(ctx, sessionID, userID, title); err == nil {
			sess.Title = title
		}
	}

	// Load messages
	dbMessages, err := s.queries.ListChatMessagesBySessionIDForUser(ctx, sessionID, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	// Build LLM messages (limit to 50)
	llmMessages := s.buildLLMMessages(dbMessages)

	var llmOpts *LLMOptions
	if s.cfg.DefaultAIAPIKey != "" && apiKey == s.cfg.DefaultAIAPIKey {
		llmOpts = &LLMOptions{Endpoint: s.cfg.DefaultAIEndpoint, Model: s.cfg.DefaultAIModel}
	} else {
		switch provider {
		case "anthropic":
			llmOpts = &LLMOptions{Endpoint: s.cfg.AnthropicEndpoint, Model: s.cfg.AnthropicModel}
		default:
			llmOpts = &LLMOptions{Endpoint: s.cfg.OpenAIEndpoint, Model: s.cfg.OpenAIModel}
		}
	}
	llm := NewLLMProvider(provider, apiKey, llmOpts)
	ch := make(chan SSEEvent, 32)

	go s.streamChat(ctx, ch, llm, llmMessages, provider, sessionID, userID)

	return ch, nil
}

func (s *ChatService) streamChat(ctx context.Context, out chan<- SSEEvent, llm LLMProvider, messages []LLMMessage, provider string, sessionID, userID uuid.UUID) {
	defer close(out)

	for round := 0; round < maxToolCalls+1; round++ {
		stream, err := llm.ChatStream(ctx, messages, KonbuTools)
		if err != nil {
			out <- SSEEvent{Event: "error", Data: map[string]string{"message": err.Error()}}
			return
		}

		var fullContent string
		var toolCalls []ToolCall
		var usage *Usage

		for ev := range stream {
			switch ev.Type {
			case "text_delta":
				fullContent += ev.Delta
				if ev.Delta != "" {
					out <- SSEEvent{Event: "text_delta", Data: map[string]string{"content": ev.Delta}}
				}
				if ev.Usage != nil {
					usage = ev.Usage
				}
			case "tool_call":
				if ev.ToolCall != nil {
					toolCalls = append(toolCalls, *ev.ToolCall)
					out <- SSEEvent{Event: "tool_call", Data: map[string]string{"tool_name": ev.ToolCall.Name}}
				}
			case "error":
				out <- SSEEvent{Event: "error", Data: map[string]string{"message": ev.Error}}
				return
			}
		}

		// No tool calls → done
		if len(toolCalls) == 0 {
			s.saveAssistantMessage(ctx, sessionID, fullContent, nil, provider, usage)
			out <- SSEEvent{Event: "done", Data: map[string]interface{}{"done": true, "session_id": sessionID.String()}}
			return
		}

		// Save assistant message with tool calls
		tcJSON, _ := json.Marshal(toolCalls)
		s.saveAssistantMessage(ctx, sessionID, fullContent, tcJSON, provider, usage)

		// Add assistant message to context
		messages = append(messages, LLMMessage{
			Role:      "assistant",
			Content:   fullContent,
			ToolCalls: toolCalls,
		})

		// Execute each tool call
		for _, tc := range toolCalls {
			result, execErr := s.executor.Execute(ctx, userID, tc.Name, tc.Arguments)
			if execErr != nil {
				result = fmt.Sprintf(`{"error": "%s"}`, execErr.Error())
			}

			out <- SSEEvent{Event: "tool_result", Data: map[string]interface{}{"tool_call_id": tc.ID, "tool_result": result}}

			// Save tool result
			tcID := tc.ID
			s.queries.CreateChatMessage(ctx, sessionID, "tool", result, nil, &tcID, nil, nil, nil, nil)

			messages = append(messages, LLMMessage{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	out <- SSEEvent{Event: "error", Data: map[string]string{"message": "too many tool call rounds"}}
}

func (s *ChatService) saveAssistantMessage(ctx context.Context, sessionID uuid.UUID, content string, toolCalls json.RawMessage, provider string, usage *Usage) {
	var inputTokens, outputTokens *int
	if usage != nil {
		inputTokens = &usage.InputTokens
		outputTokens = &usage.OutputTokens
	}
	model := providerModel(provider)
	_, err := s.queries.CreateChatMessage(ctx, sessionID, "assistant", content, toolCalls, nil, &provider, &model, inputTokens, outputTokens)
	if err != nil {
		log.Printf("failed to save assistant message: %v", err)
	}
}

func providerModel(provider string) string {
	switch provider {
	case "anthropic":
		return "claude-sonnet-4-20250514"
	default:
		return "gpt-4o"
	}
}

func (s *ChatService) buildLLMMessages(dbMessages []repository.ChatMessage) []LLMMessage {
	// System prompt
	now := time.Now().Format("2006-01-02 15:04:05 MST")
	systemPrompt := fmt.Sprintf(`あなたはkonbuのAIアシスタントです。ユーザーのメモ、ToDo、カレンダーの操作を手伝います。
ツールを使ってデータの検索・作成・更新・削除ができます。
web_searchツールでインターネット検索、web_fetchツールでWebページの内容取得ができます。
ユーザーの意図を理解し、適切なツールを呼び出してください。
ユーザーの言語に合わせて応答してください。
現在の日時: %s`, now)

	msgs := []LLMMessage{{Role: "system", Content: systemPrompt}}

	// Limit to latest 50
	start := 0
	if len(dbMessages) > 50 {
		start = len(dbMessages) - 50
	}
	for _, m := range dbMessages[start:] {
		msg := LLMMessage{Role: m.Role, Content: m.Content}
		if m.ToolCallID != nil {
			msg.ToolCallID = *m.ToolCallID
		}
		if len(m.ToolCalls) > 0 {
			var tcs []ToolCall
			json.Unmarshal(m.ToolCalls, &tcs)
			msg.ToolCalls = tcs
		}
		msgs = append(msgs, msg)
	}

	return msgs
}

func (s *ChatService) getAIKey(ctx context.Context, userID uuid.UUID) (string, string, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", apperror.Internal(err)
	}

	var settings map[string]interface{}
	if user.UserSettings != nil {
		json.Unmarshal(user.UserSettings, &settings)
	}

	provider, _ := settings["ai_provider"].(string)
	if provider == "" {
		provider = "openai"
	}

	aiKeys, _ := settings["ai_keys"].(map[string]interface{})
	if aiKeys != nil {
		encrypted, _ := aiKeys[provider].(string)
		if encrypted != "" {
			apiKey, err := Decrypt(encrypted, s.cfg.AIEncryptionKey)
			if err == nil {
				return provider, apiKey, nil
			}
		}
	}

	// Fallback to server default key
	if s.cfg.DefaultAIAPIKey != "" {
		return s.cfg.DefaultAIProvider, s.cfg.DefaultAIAPIKey, nil
	}

	return "", "", apperror.BadRequest("AI API key is not configured. Please set it in Settings.")
}

func (s *ChatService) GetAIConfig(ctx context.Context, userID uuid.UUID) (*model.AIChatConfig, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	var settings map[string]interface{}
	if user.UserSettings != nil {
		json.Unmarshal(user.UserSettings, &settings)
	}

	cfg := &model.AIChatConfig{}
	cfg.Provider, _ = settings["ai_provider"].(string)
	if cfg.Provider == "" {
		cfg.Provider = "openai"
	}

	aiKeys, _ := settings["ai_keys"].(map[string]interface{})
	if aiKeys != nil {
		if k, ok := aiKeys["openai"].(string); ok && k != "" {
			decrypted, err := Decrypt(k, s.cfg.AIEncryptionKey)
			if err == nil {
				cfg.OpenAIKeyMasked = MaskAPIKey(decrypted)
			}
		}
		if k, ok := aiKeys["anthropic"].(string); ok && k != "" {
			decrypted, err := Decrypt(k, s.cfg.AIEncryptionKey)
			if err == nil {
				cfg.AnthropicKeyMasked = MaskAPIKey(decrypted)
			}
		}
	}

	return cfg, nil
}

func (s *ChatService) SaveAIConfig(ctx context.Context, userID uuid.UUID, req model.AIChatConfig) error {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return apperror.Internal(err)
	}

	var settings map[string]interface{}
	if user.UserSettings != nil {
		json.Unmarshal(user.UserSettings, &settings)
	}
	if settings == nil {
		settings = make(map[string]interface{})
	}

	settings["ai_provider"] = req.Provider

	aiKeys, _ := settings["ai_keys"].(map[string]interface{})
	if aiKeys == nil {
		aiKeys = make(map[string]interface{})
	}

	if req.OpenAIKey != "" {
		enc, err := Encrypt(req.OpenAIKey, s.cfg.AIEncryptionKey)
		if err != nil {
			return apperror.Internal(err)
		}
		aiKeys["openai"] = enc
	}
	if req.AnthropicKey != "" {
		enc, err := Encrypt(req.AnthropicKey, s.cfg.AIEncryptionKey)
		if err != nil {
			return apperror.Internal(err)
		}
		aiKeys["anthropic"] = enc
	}

	settings["ai_keys"] = aiKeys

	settingsJSON, _ := json.Marshal(settings)
	return s.queries.UpdateUserSettings(ctx, userID, settingsJSON)
}

func toModelChatMessage(m repository.ChatMessage) model.ChatMessage {
	msg := model.ChatMessage{
		ID:        m.ID,
		Role:      m.Role,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
	if m.ToolCallID != nil {
		msg.ToolCallID = *m.ToolCallID
	}
	if len(m.ToolCalls) > 0 {
		tc := json.RawMessage(m.ToolCalls)
		msg.ToolCalls = &tc
	}
	return msg
}
