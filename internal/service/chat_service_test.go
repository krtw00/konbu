package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/repository"
)

func TestProviderModel(t *testing.T) {
	if got := providerModel("anthropic"); got != "claude-sonnet-4-20250514" {
		t.Fatalf("unexpected anthropic model: %s", got)
	}
	if got := providerModel("openai"); got != "gpt-4o" {
		t.Fatalf("unexpected default model: %s", got)
	}
}

func TestBuildLLMMessagesKeepsLatest50AndParsesToolCalls(t *testing.T) {
	svc := &ChatService{}
	now := time.Now()

	toolCalls, err := json.Marshal([]ToolCall{{
		ID:        "call_1",
		Name:      "search",
		Arguments: `{"query":"konbu"}`,
	}})
	if err != nil {
		t.Fatalf("marshal tool calls: %v", err)
	}

	var dbMessages []repository.ChatMessage
	for i := 0; i < 55; i++ {
		msg := repository.ChatMessage{
			ID:        uuid.New(),
			SessionID: uuid.New(),
			Role:      "user",
			Content:   fmt.Sprintf("message-%d", i),
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
		}
		if i == 54 {
			msg.Role = "assistant"
			msg.ToolCalls = toolCalls
		}
		dbMessages = append(dbMessages, msg)
	}

	got := svc.buildLLMMessages(dbMessages)
	if len(got) != 51 {
		t.Fatalf("expected system prompt + 50 messages, got %d", len(got))
	}
	if got[0].Role != "system" || !strings.Contains(got[0].Content, "konbuのAIアシスタント") {
		t.Fatalf("unexpected system prompt: %#v", got[0])
	}
	if got[1].Content != "message-5" {
		t.Fatalf("expected oldest retained message to be message-5, got %q", got[1].Content)
	}
	last := got[len(got)-1]
	if last.Role != "assistant" || len(last.ToolCalls) != 1 || last.ToolCalls[0].Name != "search" {
		t.Fatalf("unexpected last message: %#v", last)
	}
}
