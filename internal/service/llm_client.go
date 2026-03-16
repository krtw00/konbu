package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type LLMProvider interface {
	ChatStream(ctx context.Context, messages []LLMMessage, tools []ToolDefinition) (<-chan StreamEvent, error)
}

type LLMMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type StreamEvent struct {
	Type     string
	Delta    string
	ToolCall *ToolCall
	Usage    *Usage
	Error    string
}

type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type LLMOptions struct {
	Endpoint string
	Model    string
}

func NewLLMProvider(provider, apiKey string, opts *LLMOptions) LLMProvider {
	switch provider {
	case "anthropic":
		endpoint := "https://api.anthropic.com/v1/messages"
		model := "claude-sonnet-4-20250514"
		if opts != nil {
			if opts.Endpoint != "" {
				endpoint = opts.Endpoint
			}
			if opts.Model != "" {
				model = opts.Model
			}
		}
		return &anthropicProvider{apiKey: apiKey, endpoint: endpoint, model: model, client: &http.Client{}}
	default:
		endpoint := "https://api.openai.com/v1/chat/completions"
		model := "gpt-4o"
		if opts != nil {
			if opts.Endpoint != "" {
				endpoint = opts.Endpoint
			}
			if opts.Model != "" {
				model = opts.Model
			}
		}
		return &openaiProvider{apiKey: apiKey, endpoint: endpoint, model: model, client: &http.Client{}}
	}
}

// --- OpenAI ---

type openaiProvider struct {
	apiKey   string
	endpoint string
	model    string
	client   *http.Client
}

type openaiRequest struct {
	Model    string              `json:"model"`
	Messages []openaiMessage     `json:"messages"`
	Stream   bool                `json:"stream"`
	Tools    []openaiTool        `json:"tools,omitempty"`
	StreamOptions *openaiStreamOpts `json:"stream_options,omitempty"`
}

type openaiStreamOpts struct {
	IncludeUsage bool `json:"include_usage"`
}

type openaiMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openaiTool struct {
	Type     string         `json:"type"`
	Function openaiFunction `json:"function"`
}

type openaiFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type openaiToolCall struct {
	Index    int                  `json:"index"`
	ID       string               `json:"id,omitempty"`
	Type     string               `json:"type,omitempty"`
	Function openaiToolCallFunc   `json:"function"`
}

type openaiToolCallFunc struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

func (p *openaiProvider) ChatStream(ctx context.Context, messages []LLMMessage, tools []ToolDefinition) (<-chan StreamEvent, error) {
	oaiMessages := make([]openaiMessage, 0, len(messages))
	for _, m := range messages {
		msg := openaiMessage{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		for _, tc := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, openaiToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: openaiToolCallFunc{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			})
		}
		oaiMessages = append(oaiMessages, msg)
	}

	var oaiTools []openaiTool
	for _, t := range tools {
		oaiTools = append(oaiTools, openaiTool{
			Type: "function",
			Function: openaiFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	reqBody := openaiRequest{
		Model:    p.model,
		Messages: oaiMessages,
		Stream:   true,
		Tools:    oaiTools,
		StreamOptions: &openaiStreamOpts{IncludeUsage: true},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai api error (status %d): %s", resp.StatusCode, string(errBody))
	}

	ch := make(chan StreamEvent, 32)
	go p.parseSSE(resp, ch)
	return ch, nil
}

func (p *openaiProvider) parseSSE(resp *http.Response, ch chan<- StreamEvent) {
	defer resp.Body.Close()
	defer close(ch)

	// tool_calls arrive in fragments; buffer per index
	toolBuf := make(map[int]*ToolCall)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			// flush remaining tool calls
			for _, tc := range toolBuf {
				ch <- StreamEvent{Type: "tool_call", ToolCall: tc}
			}
			ch <- StreamEvent{Type: "done"}
			return
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content   string           `json:"content"`
					ToolCalls []openaiToolCall  `json:"tool_calls"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			ch <- StreamEvent{Type: "error", Error: fmt.Sprintf("parse chunk: %v", err)}
			return
		}

		if chunk.Usage != nil {
			ch <- StreamEvent{
				Type: "text_delta",
				Usage: &Usage{
					InputTokens:  chunk.Usage.PromptTokens,
					OutputTokens: chunk.Usage.CompletionTokens,
				},
			}
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta

		if delta.Content != "" {
			ch <- StreamEvent{Type: "text_delta", Delta: delta.Content}
		}

		for _, tc := range delta.ToolCalls {
			existing, ok := toolBuf[tc.Index]
			if !ok {
				existing = &ToolCall{}
				toolBuf[tc.Index] = existing
			}
			if tc.ID != "" {
				existing.ID = tc.ID
			}
			if tc.Function.Name != "" {
				existing.Name = tc.Function.Name
			}
			existing.Arguments += tc.Function.Arguments
		}
	}

	if err := scanner.Err(); err != nil {
		ch <- StreamEvent{Type: "error", Error: fmt.Sprintf("read stream: %v", err)}
		return
	}

	// flush if stream ended without [DONE]
	for _, tc := range toolBuf {
		ch <- StreamEvent{Type: "tool_call", ToolCall: tc}
	}
	ch <- StreamEvent{Type: "done"}
}

// --- Anthropic ---

type anthropicProvider struct {
	apiKey   string
	endpoint string
	model    string
	client   *http.Client
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Stream    bool               `json:"stream"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type anthropicContentBlock struct {
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Input       json.RawMessage `json:"input,omitempty"`
	ToolUseID   string `json:"tool_use_id,omitempty"`
	Content     string `json:"content,omitempty"`
}

type anthropicTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
}

func (p *anthropicProvider) ChatStream(ctx context.Context, messages []LLMMessage, tools []ToolDefinition) (<-chan StreamEvent, error) {
	var systemPrompt string
	var antMessages []anthropicMessage

	for _, m := range messages {
		switch m.Role {
		case "system":
			systemPrompt = m.Content
		case "assistant":
			var content []anthropicContentBlock
			if m.Content != "" {
				content = append(content, anthropicContentBlock{Type: "text", Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				input := json.RawMessage(tc.Arguments)
				if len(input) == 0 {
					input = json.RawMessage("{}")
				}
				content = append(content, anthropicContentBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: input,
				})
			}
			if len(content) > 0 {
				antMessages = append(antMessages, anthropicMessage{Role: "assistant", Content: content})
			}
		case "tool":
			antMessages = append(antMessages, anthropicMessage{
				Role: "user",
				Content: []anthropicContentBlock{
					{
						Type:      "tool_result",
						ToolUseID: m.ToolCallID,
						Content:   m.Content,
					},
				},
			})
		default:
			antMessages = append(antMessages, anthropicMessage{Role: m.Role, Content: m.Content})
		}
	}

	var antTools []anthropicTool
	for _, t := range tools {
		antTools = append(antTools, anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		})
	}

	reqBody := anthropicRequest{
		Model:     p.model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages:  antMessages,
		Stream:    true,
		Tools:     antTools,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic api error (status %d): %s", resp.StatusCode, string(errBody))
	}

	ch := make(chan StreamEvent, 32)
	go p.parseSSE(resp, ch)
	return ch, nil
}

func (p *anthropicProvider) parseSSE(resp *http.Response, ch chan<- StreamEvent) {
	defer resp.Body.Close()
	defer close(ch)

	// buffer tool_use arguments per content block index
	toolBuf := make(map[int]*ToolCall)
	var currentEvent string

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		switch currentEvent {
		case "content_block_start":
			var ev struct {
				Index        int `json:"index"`
				ContentBlock struct {
					Type string `json:"type"`
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"content_block"`
			}
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				ch <- StreamEvent{Type: "error", Error: fmt.Sprintf("parse content_block_start: %v", err)}
				return
			}
			if ev.ContentBlock.Type == "tool_use" {
				toolBuf[ev.Index] = &ToolCall{
					ID:   ev.ContentBlock.ID,
					Name: ev.ContentBlock.Name,
				}
			}

		case "content_block_delta":
			var ev struct {
				Index int `json:"index"`
				Delta struct {
					Type        string `json:"type"`
					Text        string `json:"text"`
					PartialJSON string `json:"partial_json"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				ch <- StreamEvent{Type: "error", Error: fmt.Sprintf("parse content_block_delta: %v", err)}
				return
			}
			switch ev.Delta.Type {
			case "text_delta":
				ch <- StreamEvent{Type: "text_delta", Delta: ev.Delta.Text}
			case "input_json_delta":
				if tc, ok := toolBuf[ev.Index]; ok {
					tc.Arguments += ev.Delta.PartialJSON
				}
			}

		case "content_block_stop":
			var ev struct {
				Index int `json:"index"`
			}
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				continue
			}
			if tc, ok := toolBuf[ev.Index]; ok {
				ch <- StreamEvent{Type: "tool_call", ToolCall: tc}
				delete(toolBuf, ev.Index)
			}

		case "message_delta":
			var ev struct {
				Usage struct {
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			}
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				continue
			}
			ch <- StreamEvent{Type: "text_delta", Usage: &Usage{OutputTokens: ev.Usage.OutputTokens}}

		case "message_start":
			var ev struct {
				Message struct {
					Usage struct {
						InputTokens int `json:"input_tokens"`
					} `json:"usage"`
				} `json:"message"`
			}
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				continue
			}
			if ev.Message.Usage.InputTokens > 0 {
				ch <- StreamEvent{Type: "text_delta", Usage: &Usage{InputTokens: ev.Message.Usage.InputTokens}}
			}

		case "message_stop":
			ch <- StreamEvent{Type: "done"}
			return

		case "error":
			var ev struct {
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				ch <- StreamEvent{Type: "error", Error: data}
			} else {
				ch <- StreamEvent{Type: "error", Error: ev.Error.Message}
			}
			return
		}

		currentEvent = ""
	}

	if err := scanner.Err(); err != nil {
		ch <- StreamEvent{Type: "error", Error: fmt.Sprintf("read stream: %v", err)}
		return
	}
}
