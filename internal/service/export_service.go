package service

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/model"
)

type ExportData struct {
	Version    int                  `json:"version"`
	ExportedAt time.Time            `json:"exported_at"`
	Memos      []model.Memo         `json:"memos"`
	Todos      []model.Todo         `json:"todos"`
	Events     []model.CalendarEvent `json:"events"`
	Tools      []model.Tool         `json:"tools"`
}

type ExportService struct {
	memoSvc  *MemoService
	todoSvc  *TodoService
	eventSvc *EventService
	toolSvc  *ToolService
}

func NewExportService(db *sql.DB, memoSvc *MemoService, todoSvc *TodoService, eventSvc *EventService, toolSvc *ToolService) *ExportService {
	return &ExportService{
		memoSvc:  memoSvc,
		todoSvc:  todoSvc,
		eventSvc: eventSvc,
		toolSvc:  toolSvc,
	}
}

func (s *ExportService) allParams() model.ListParams {
	return model.ListParams{Limit: 10000, Offset: 0}
}

func (s *ExportService) fetchAllData(ctx context.Context, userID uuid.UUID) (*ExportData, error) {
	params := s.allParams()

	memoResult, err := s.memoSvc.ListMemos(ctx, userID, params)
	if err != nil {
		return nil, err
	}
	memos := memoResult.Data.([]model.Memo)

	// Fetch content for each memo
	for i, m := range memos {
		full, err := s.memoSvc.GetMemo(ctx, m.ID, userID)
		if err != nil {
			return nil, err
		}
		memos[i] = *full
	}

	todoResult, err := s.todoSvc.ListTodos(ctx, userID, params)
	if err != nil {
		return nil, err
	}
	todos := todoResult.Data.([]model.Todo)

	eventResult, err := s.eventSvc.ListEvents(ctx, userID, params)
	if err != nil {
		return nil, err
	}
	events := eventResult.Data.([]model.CalendarEvent)

	tools, err := s.toolSvc.ListTools(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &ExportData{
		Version:    1,
		ExportedAt: time.Now().UTC(),
		Memos:      memos,
		Todos:      todos,
		Events:     events,
		Tools:      tools,
	}, nil
}

func (s *ExportService) ExportJSON(ctx context.Context, userID uuid.UUID) (*ExportData, error) {
	return s.fetchAllData(ctx, userID)
}

func (s *ExportService) ExportMarkdown(ctx context.Context, userID uuid.UUID) (*bytes.Buffer, error) {
	data, err := s.fetchAllData(ctx, userID)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	for _, m := range data.Memos {
		filename := fmt.Sprintf("memos/%s.md", sanitizeFilename(m.Title))
		fw, err := zw.Create(filename)
		if err != nil {
			return nil, apperror.Internal(err)
		}

		var sb strings.Builder
		sb.WriteString("---\n")
		sb.WriteString(fmt.Sprintf("title: %q\n", m.Title))
		if len(m.Tags) > 0 {
			tagNames := make([]string, len(m.Tags))
			for i, t := range m.Tags {
				tagNames[i] = t.Name
			}
			sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(tagNames, ", ")))
		}
		sb.WriteString(fmt.Sprintf("created_at: %s\n", m.CreatedAt.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("updated_at: %s\n", m.UpdatedAt.Format(time.RFC3339)))
		sb.WriteString("---\n\n")
		if m.Content != nil {
			sb.WriteString(*m.Content)
		}

		if _, err := fw.Write([]byte(sb.String())); err != nil {
			return nil, apperror.Internal(err)
		}
	}

	if err := s.writeTodosMarkdown(zw, data.Todos); err != nil {
		return nil, err
	}
	if err := s.writeEventsMarkdown(zw, data.Events); err != nil {
		return nil, err
	}
	if err := s.writeToolsMarkdown(zw, data.Tools); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, apperror.Internal(err)
	}

	return buf, nil
}

func (s *ExportService) writeTodosMarkdown(zw *zip.Writer, todos []model.Todo) error {
	fw, err := zw.Create("todos.md")
	if err != nil {
		return apperror.Internal(err)
	}

	var sb strings.Builder
	sb.WriteString("# Todos\n\n")
	for _, t := range todos {
		check := " "
		if t.Status == "done" {
			check = "x"
		}
		line := fmt.Sprintf("- [%s] %s", check, t.Title)
		if t.DueDate != nil {
			line += fmt.Sprintf(" (due: %s)", *t.DueDate)
		}
		if t.Description != "" {
			line += fmt.Sprintf("\n  %s", t.Description)
		}
		sb.WriteString(line + "\n")
	}

	_, err = fw.Write([]byte(sb.String()))
	if err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *ExportService) writeEventsMarkdown(zw *zip.Writer, events []model.CalendarEvent) error {
	fw, err := zw.Create("events.md")
	if err != nil {
		return apperror.Internal(err)
	}

	var sb strings.Builder
	sb.WriteString("# Events\n\n")
	sb.WriteString("| Title | Start | End | All Day | Description |\n")
	sb.WriteString("|-------|-------|-----|---------|-------------|\n")
	for _, e := range events {
		endStr := ""
		if e.EndAt != nil {
			endStr = e.EndAt.Format("2006-01-02 15:04")
		}
		allDay := "No"
		if e.AllDay {
			allDay = "Yes"
		}
		desc := strings.ReplaceAll(e.Description, "\n", " ")
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			e.Title,
			e.StartAt.Format("2006-01-02 15:04"),
			endStr,
			allDay,
			desc,
		))
	}

	_, err = fw.Write([]byte(sb.String()))
	if err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *ExportService) writeToolsMarkdown(zw *zip.Writer, tools []model.Tool) error {
	fw, err := zw.Create("tools.md")
	if err != nil {
		return apperror.Internal(err)
	}

	var sb strings.Builder
	sb.WriteString("# Tools\n\n")
	sb.WriteString("| Name | URL | Category | Icon |\n")
	sb.WriteString("|------|-----|----------|------|\n")
	for _, t := range tools {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			t.Name, t.URL, t.Category, t.Icon,
		))
	}

	_, err = fw.Write([]byte(sb.String()))
	if err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	s := replacer.Replace(name)
	if s == "" {
		s = "untitled"
	}
	return s
}
