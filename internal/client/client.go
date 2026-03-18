package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{BaseURL: baseURL, APIKey: apiKey, HTTPClient: &http.Client{}}
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
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
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

func (c *Client) doRaw(method, path string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	return c.HTTPClient.Do(req)
}

// --- Memo ---

type Memo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	Tags      []Tag  `json:"tags"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type MemoRow struct {
	ID        string            `json:"id"`
	MemoID    string            `json:"memo_id"`
	RowData   map[string]string `json:"row_data"`
	SortOrder int               `json:"sort_order"`
}

type MemoRowsResponse struct {
	Data  []MemoRow `json:"data"`
	Total int       `json:"total"`
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

func (c *Client) UpdateMemo(id string, fields map[string]any) (*Memo, error) {
	data, err := c.do("PUT", "/api/v1/memos/"+id, fields)
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

func (c *Client) ListMemoRows(memoID string) ([]MemoRow, error) {
	data, err := c.do("GET", "/api/v1/memos/"+memoID+"/rows?limit=100", nil)
	if err != nil {
		return nil, err
	}
	var rows []MemoRow
	if err := json.Unmarshal(data, &rows); err == nil {
		return rows, nil
	}
	var resp MemoRowsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) CreateMemoRow(memoID string, rowData map[string]string) (*MemoRow, error) {
	data, err := c.do("POST", "/api/v1/memos/"+memoID+"/rows", map[string]any{
		"row_data": rowData,
	})
	if err != nil {
		return nil, err
	}
	var row MemoRow
	return &row, json.Unmarshal(data, &row)
}

func (c *Client) DeleteMemoRow(memoID, rowID string) error {
	_, err := c.do("DELETE", "/api/v1/memos/"+memoID+"/rows/"+rowID, nil)
	return err
}

func (c *Client) ExportMemoRowsCSV(memoID string) (string, error) {
	resp, err := c.doRaw("GET", "/api/v1/memos/"+memoID+"/rows/export")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

// --- Todo ---

type Todo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	DueDate     string `json:"due_date"`
	Tags        []Tag  `json:"tags"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (c *Client) ListTodos() ([]Todo, error) {
	data, err := c.do("GET", "/api/v1/todos", nil)
	if err != nil {
		return nil, err
	}
	var todos []Todo
	return todos, json.Unmarshal(data, &todos)
}

func (c *Client) GetTodo(id string) (*Todo, error) {
	data, err := c.do("GET", "/api/v1/todos/"+id, nil)
	if err != nil {
		return nil, err
	}
	var t Todo
	return &t, json.Unmarshal(data, &t)
}

func (c *Client) CreateTodo(fields map[string]any) (*Todo, error) {
	if _, ok := fields["status"]; !ok {
		fields["status"] = "open"
	}
	if _, ok := fields["tags"]; !ok {
		fields["tags"] = []string{}
	}
	data, err := c.do("POST", "/api/v1/todos", fields)
	if err != nil {
		return nil, err
	}
	var t Todo
	return &t, json.Unmarshal(data, &t)
}

func (c *Client) UpdateTodo(id string, fields map[string]any) (*Todo, error) {
	data, err := c.do("PUT", "/api/v1/todos/"+id, fields)
	if err != nil {
		return nil, err
	}
	var t Todo
	return &t, json.Unmarshal(data, &t)
}

func (c *Client) DoneTodo(id string) error {
	_, err := c.do("PATCH", "/api/v1/todos/"+id+"/done", nil)
	return err
}

func (c *Client) ReopenTodo(id string) error {
	_, err := c.do("PATCH", "/api/v1/todos/"+id+"/reopen", nil)
	return err
}

func (c *Client) DeleteTodo(id string) error {
	_, err := c.do("DELETE", "/api/v1/todos/"+id, nil)
	return err
}

// --- Event ---

type Event struct {
	CalendarID     *string `json:"calendar_id"`
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	StartAt        string  `json:"start_at"`
	EndAt          *string `json:"end_at"`
	AllDay         bool    `json:"all_day"`
	RecurrenceRule *string `json:"recurrence_rule"`
	Tags           []Tag   `json:"tags"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

func (c *Client) ListEvents(calendarID string) ([]Event, error) {
	path := "/api/v1/events?limit=100&sort=start_at:asc"
	if calendarID != "" {
		path += "&calendar_id=" + url.QueryEscape(calendarID)
	}
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var events []Event
	return events, json.Unmarshal(data, &events)
}

func (c *Client) GetEvent(id string) (*Event, error) {
	data, err := c.do("GET", "/api/v1/events/"+id, nil)
	if err != nil {
		return nil, err
	}
	var e Event
	return &e, json.Unmarshal(data, &e)
}

func (c *Client) CreateEvent(fields map[string]any) (*Event, error) {
	data, err := c.do("POST", "/api/v1/events", fields)
	if err != nil {
		return nil, err
	}
	var e Event
	return &e, json.Unmarshal(data, &e)
}

func (c *Client) UpdateEvent(id string, fields map[string]any) (*Event, error) {
	data, err := c.do("PUT", "/api/v1/events/"+id, fields)
	if err != nil {
		return nil, err
	}
	var e Event
	return &e, json.Unmarshal(data, &e)
}

func (c *Client) DeleteEvent(id string) error {
	_, err := c.do("DELETE", "/api/v1/events/"+id, nil)
	return err
}

// --- Calendar ---

type Calendar struct {
	ID          string  `json:"id"`
	OwnerID     string  `json:"owner_id"`
	Name        string  `json:"name"`
	IsDefault   bool    `json:"is_default"`
	ShareToken  *string `json:"share_token"`
	Color       string  `json:"color"`
	MemberCount int     `json:"member_count"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CalendarMember struct {
	CalendarID string `json:"calendar_id"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
	Role       string `json:"role"`
	Color      string `json:"color"`
	JoinedAt   string `json:"joined_at"`
}

type CalendarDetail struct {
	Calendar
	Members []CalendarMember `json:"members"`
}

func (c *Client) ListCalendars() ([]Calendar, error) {
	data, err := c.do("GET", "/api/v1/calendars", nil)
	if err != nil {
		return nil, err
	}
	var calendars []Calendar
	return calendars, json.Unmarshal(data, &calendars)
}

func (c *Client) GetCalendar(id string) (*CalendarDetail, error) {
	data, err := c.do("GET", "/api/v1/calendars/"+id, nil)
	if err != nil {
		return nil, err
	}
	var detail CalendarDetail
	return &detail, json.Unmarshal(data, &detail)
}

func (c *Client) CreateCalendar(name, color string) (*Calendar, error) {
	body := map[string]any{"name": name}
	if color != "" {
		body["color"] = color
	}
	data, err := c.do("POST", "/api/v1/calendars", body)
	if err != nil {
		return nil, err
	}
	var cal Calendar
	return &cal, json.Unmarshal(data, &cal)
}

func (c *Client) UpdateCalendar(id string, fields map[string]any) (*Calendar, error) {
	data, err := c.do("PUT", "/api/v1/calendars/"+id, fields)
	if err != nil {
		return nil, err
	}
	var cal Calendar
	return &cal, json.Unmarshal(data, &cal)
}

func (c *Client) DeleteCalendar(id string) error {
	_, err := c.do("DELETE", "/api/v1/calendars/"+id, nil)
	return err
}

func (c *Client) JoinCalendar(token string) (*Calendar, error) {
	data, err := c.do("POST", "/api/v1/calendars/join/"+token, nil)
	if err != nil {
		return nil, err
	}
	var cal Calendar
	return &cal, json.Unmarshal(data, &cal)
}

func (c *Client) CreateCalendarShareLink(id string) (string, error) {
	data, err := c.do("POST", "/api/v1/calendars/"+id+"/share-link", nil)
	if err != nil {
		return "", err
	}
	var resp struct {
		ShareToken string `json:"share_token"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.ShareToken, nil
}

func (c *Client) DeleteCalendarShareLink(id string) error {
	_, err := c.do("DELETE", "/api/v1/calendars/"+id+"/share-link", nil)
	return err
}

func (c *Client) AddCalendarMember(id, email, role string) (*CalendarMember, error) {
	data, err := c.do("POST", "/api/v1/calendars/"+id+"/members", map[string]any{
		"email": email,
		"role":  role,
	})
	if err != nil {
		return nil, err
	}
	var member CalendarMember
	return &member, json.Unmarshal(data, &member)
}

func (c *Client) UpdateCalendarMember(id, userID string, fields map[string]any) error {
	_, err := c.do("PUT", "/api/v1/calendars/"+id+"/members/"+userID, fields)
	return err
}

func (c *Client) RemoveCalendarMember(id, userID string) error {
	_, err := c.do("DELETE", "/api/v1/calendars/"+id+"/members/"+userID, nil)
	return err
}

// --- Public Share ---

type PublicShare struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Token        string `json:"token"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func (c *Client) GetPublicShare(resourceType, id string) (*PublicShare, error) {
	data, err := c.do("GET", "/api/v1/public-shares/"+resourceType+"/"+id, nil)
	if err != nil {
		return nil, err
	}
	if string(data) == "null" {
		return nil, nil
	}
	var share PublicShare
	return &share, json.Unmarshal(data, &share)
}

func (c *Client) CreatePublicShare(resourceType, id string) (*PublicShare, error) {
	data, err := c.do("POST", "/api/v1/public-shares/"+resourceType+"/"+id, nil)
	if err != nil {
		return nil, err
	}
	var share PublicShare
	return &share, json.Unmarshal(data, &share)
}

func (c *Client) DeletePublicShare(resourceType, id string) error {
	_, err := c.do("DELETE", "/api/v1/public-shares/"+resourceType+"/"+id, nil)
	return err
}

func (c *Client) PublicURL(token string) string {
	return c.BaseURL + "/public/" + token
}

// --- Tool ---

type Tool struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Icon     string `json:"icon"`
	Category string `json:"category"`
}

func (c *Client) ListTools() ([]Tool, error) {
	data, err := c.do("GET", "/api/v1/tools", nil)
	if err != nil {
		return nil, err
	}
	var tools []Tool
	return tools, json.Unmarshal(data, &tools)
}

func (c *Client) CreateTool(name, url, category string) (*Tool, error) {
	body := map[string]any{"name": name, "url": url}
	if category != "" {
		body["category"] = category
	}
	data, err := c.do("POST", "/api/v1/tools", body)
	if err != nil {
		return nil, err
	}
	var t Tool
	return &t, json.Unmarshal(data, &t)
}

func (c *Client) UpdateTool(id string, fields map[string]any) (*Tool, error) {
	data, err := c.do("PUT", "/api/v1/tools/"+id, fields)
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

func (c *Client) ReorderTools(order []string) error {
	_, err := c.do("PUT", "/api/v1/tools/reorder", map[string]any{"order": order})
	return err
}

// --- Tag ---

func (c *Client) ListTags() ([]Tag, error) {
	data, err := c.do("GET", "/api/v1/tags", nil)
	if err != nil {
		return nil, err
	}
	var tags []Tag
	return tags, json.Unmarshal(data, &tags)
}

func (c *Client) DeleteTag(id string) error {
	_, err := c.do("DELETE", "/api/v1/tags/"+id, nil)
	return err
}

// --- API Key ---

type APIKey struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key,omitempty"`
	CreatedAt string `json:"created_at"`
}

type CalendarFeedTokenStatus struct {
	Exists     bool    `json:"exists"`
	CreatedAt  *string `json:"created_at,omitempty"`
	LastUsedAt *string `json:"last_used_at,omitempty"`
}

type CalendarFeedToken struct {
	Token      string  `json:"token"`
	URL        string  `json:"url"`
	CreatedAt  string  `json:"created_at"`
	LastUsedAt *string `json:"last_used_at,omitempty"`
}

func (c *Client) ListAPIKeys() ([]APIKey, error) {
	data, err := c.do("GET", "/api/v1/api-keys", nil)
	if err != nil {
		return nil, err
	}
	var keys []APIKey
	return keys, json.Unmarshal(data, &keys)
}

func (c *Client) CreateAPIKey(name string) (*APIKey, error) {
	data, err := c.do("POST", "/api/v1/api-keys", map[string]any{"name": name})
	if err != nil {
		return nil, err
	}
	var k APIKey
	return &k, json.Unmarshal(data, &k)
}

func (c *Client) DeleteAPIKey(id string) error {
	_, err := c.do("DELETE", "/api/v1/api-keys/"+id, nil)
	return err
}

func (c *Client) GetCalendarFeedTokenStatus() (*CalendarFeedTokenStatus, error) {
	data, err := c.do("GET", "/api/v1/api-keys/calendar-feed", nil)
	if err != nil {
		return nil, err
	}
	var status CalendarFeedTokenStatus
	return &status, json.Unmarshal(data, &status)
}

func (c *Client) RotateCalendarFeedToken() (*CalendarFeedToken, error) {
	data, err := c.do("POST", "/api/v1/api-keys/calendar-feed", map[string]any{})
	if err != nil {
		return nil, err
	}
	var token CalendarFeedToken
	return &token, json.Unmarshal(data, &token)
}

func (c *Client) DeleteCalendarFeedToken() error {
	_, err := c.do("DELETE", "/api/v1/api-keys/calendar-feed", nil)
	return err
}

// --- Search ---

type SearchResult struct {
	Type    string   `json:"type"`
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Snippet string   `json:"snippet"`
	Tags    []string `json:"tags"`
}

func (c *Client) Search(query string) ([]SearchResult, error) {
	data, err := c.do("GET", "/api/v1/search?q="+url.QueryEscape(query), nil)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	return results, json.Unmarshal(data, &results)
}

type AttachmentResult struct {
	URL string `json:"url"`
}

// --- Export ---

func (c *Client) ExportJSON(outPath string) error {
	resp, err := c.doRaw("GET", "/api/v1/export/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func (c *Client) ExportMarkdown(outPath string) error {
	resp, err := c.doRaw("GET", "/api/v1/export/markdown")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func (c *Client) ExportICal(outPath string) error {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/v1/calendar.ics", nil)
	if err != nil {
		return err
	}
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	if outPath == "" {
		_, err = io.Copy(os.Stdout, resp.Body)
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// --- Import ---

func (c *Client) ImportICal(filePath string) (json.RawMessage, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.BaseURL+"/api/v1/import/ical", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
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
	return raw, nil
}

func (c *Client) UploadAttachment(filePath string) (*AttachmentResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.BaseURL+"/api/v1/attachments", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
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
		return nil, err
	}
	var result AttachmentResult
	return &result, json.Unmarshal(ar.Data, &result)
}
