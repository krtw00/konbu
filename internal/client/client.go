package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{BaseURL: baseURL, HTTPClient: &http.Client{}}
}

type apiResponse struct {
	Data json.RawMessage `json:"data"`
}

func (c *Client) do(method, path string, body any) (json.RawMessage, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, r)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var ar apiResponse
	if err := json.Unmarshal(raw, &ar); err != nil {
		return raw, nil
	}
	return ar.Data, nil
}

// Memo

type Memo struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Type      string   `json:"type"`
	Tags      []Tag    `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type Tag struct {
	Name string `json:"name"`
}

func (c *Client) ListMemos() ([]Memo, error) {
	data, err := c.do("GET", "/api/v1/memos", nil)
	if err != nil {
		return nil, err
	}
	var memos []Memo
	return memos, json.Unmarshal(data, &memos)
}

func (c *Client) CreateMemo(title, content string, tags []string) (*Memo, error) {
	data, err := c.do("POST", "/api/v1/memos", map[string]any{
		"title": title, "content": content, "type": "markdown", "tags": tags,
	})
	if err != nil {
		return nil, err
	}
	var m Memo
	return &m, json.Unmarshal(data, &m)
}

func (c *Client) GetMemo(id string) (*Memo, error) {
	data, err := c.do("GET", "/api/v1/memos/"+id, nil)
	if err != nil {
		return nil, err
	}
	var m Memo
	return &m, json.Unmarshal(data, &m)
}

func (c *Client) DeleteMemo(id string) error {
	_, err := c.do("DELETE", "/api/v1/memos/"+id, nil)
	return err
}

// Todo

type Todo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	DueDate     string `json:"due_date"`
	Tags        []Tag  `json:"tags"`
	CreatedAt   string `json:"created_at"`
}

func (c *Client) ListTodos() ([]Todo, error) {
	data, err := c.do("GET", "/api/v1/todos", nil)
	if err != nil {
		return nil, err
	}
	var todos []Todo
	return todos, json.Unmarshal(data, &todos)
}

func (c *Client) CreateTodo(title string, tags []string) (*Todo, error) {
	body := map[string]any{"title": title, "status": "open", "tags": tags}
	data, err := c.do("POST", "/api/v1/todos", body)
	if err != nil {
		return nil, err
	}
	var t Todo
	return &t, json.Unmarshal(data, &t)
}

func (c *Client) UpdateTodo(id string, fields map[string]any) error {
	_, err := c.do("PATCH", "/api/v1/todos/"+id, fields)
	return err
}

func (c *Client) DeleteTodo(id string) error {
	_, err := c.do("DELETE", "/api/v1/todos/"+id, nil)
	return err
}

// Tool

type Tool struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon string `json:"icon"`
}

func (c *Client) ListTools() ([]Tool, error) {
	data, err := c.do("GET", "/api/v1/tools", nil)
	if err != nil {
		return nil, err
	}
	var tools []Tool
	return tools, json.Unmarshal(data, &tools)
}

func (c *Client) CreateTool(name, url string) (*Tool, error) {
	data, err := c.do("POST", "/api/v1/tools", map[string]any{"name": name, "url": url})
	if err != nil {
		return nil, err
	}
	var t Tool
	return &t, json.Unmarshal(data, &t)
}

func (c *Client) DeleteTool(id string) error {
	_, err := c.do("DELETE", "/api/v1/tools/"+id, nil)
	return err
}
