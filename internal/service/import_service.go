package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/model"
)

type ImportService struct {
	eventSvc *EventService
}

func NewImportService(eventSvc *EventService) *ImportService {
	return &ImportService{eventSvc: eventSvc}
}

type icalEvent struct {
	summary     string
	description string
	dtstart     string
	dtend       string
	rrule       string
}

func (s *ImportService) ImportICal(ctx context.Context, userID uuid.UUID, r io.Reader) ([]*model.CalendarEvent, error) {
	rawEvents, err := parseICal(r)
	if err != nil {
		return nil, fmt.Errorf("ical parse error: %w", err)
	}

	var created []*model.CalendarEvent
	for _, ev := range rawEvents {
		req, err := toCreateEventRequest(ev)
		if err != nil {
			return nil, fmt.Errorf("event conversion error: %w", err)
		}

		event, err := s.eventSvc.CreateEvent(ctx, userID, req)
		if err != nil {
			return nil, fmt.Errorf("create event error: %w", err)
		}
		created = append(created, event)
	}

	return created, nil
}

func parseICal(r io.Reader) ([]icalEvent, error) {
	scanner := bufio.NewScanner(r)
	var events []icalEvent
	var current *icalEvent
	var lastKey string

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")

		// RFC 5545: continuation lines start with space or tab
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if current != nil {
				val := line[1:]
				switch lastKey {
				case "SUMMARY":
					current.summary += val
				case "DESCRIPTION":
					current.description += val
				case "DTSTART":
					current.dtstart += val
				case "DTEND":
					current.dtend += val
				case "RRULE":
					current.rrule += val
				}
			}
			continue
		}

		if line == "BEGIN:VEVENT" {
			current = &icalEvent{}
			lastKey = ""
			continue
		}
		if line == "END:VEVENT" {
			if current != nil {
				events = append(events, *current)
				current = nil
			}
			continue
		}

		if current == nil {
			continue
		}

		key, val := splitProperty(line)
		// Strip parameters (e.g. DTSTART;VALUE=DATE:20240101 → key=DTSTART)
		baseKey := key
		if idx := strings.Index(baseKey, ";"); idx >= 0 {
			baseKey = baseKey[:idx]
		}

		switch baseKey {
		case "SUMMARY":
			current.summary = val
			lastKey = "SUMMARY"
		case "DESCRIPTION":
			current.description = unescapeICal(val)
			lastKey = "DESCRIPTION"
		case "DTSTART":
			current.dtstart = val
			lastKey = "DTSTART"
		case "DTEND":
			current.dtend = val
			lastKey = "DTEND"
		case "RRULE":
			current.rrule = val
			lastKey = "RRULE"
		default:
			lastKey = ""
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func splitProperty(line string) (string, string) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return line, ""
	}
	return line[:idx], line[idx+1:]
}

func unescapeICal(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\,`, ",")
	s = strings.ReplaceAll(s, `\\`, `\`)
	return s
}

func toCreateEventRequest(ev icalEvent) (model.CreateEventRequest, error) {
	startAt, allDay, err := parseICalDateTime(ev.dtstart)
	if err != nil {
		return model.CreateEventRequest{}, fmt.Errorf("invalid DTSTART %q: %w", ev.dtstart, err)
	}

	req := model.CreateEventRequest{
		Title:       ev.summary,
		Description: ev.description,
		StartAt:     startAt,
		AllDay:      allDay,
	}

	if ev.dtend != "" {
		endAt, _, err := parseICalDateTime(ev.dtend)
		if err != nil {
			return model.CreateEventRequest{}, fmt.Errorf("invalid DTEND %q: %w", ev.dtend, err)
		}
		req.EndAt = &endAt
	}

	if ev.rrule != "" {
		if rule := mapRRule(ev.rrule); rule != "" {
			req.RecurrenceRule = &rule
		}
	}

	return req, nil
}

func parseICalDateTime(s string) (time.Time, bool, error) {
	s = strings.TrimSpace(s)

	// Date only: YYYYMMDD → all_day event
	if len(s) == 8 {
		t, err := time.Parse("20060102", s)
		if err != nil {
			return time.Time{}, false, err
		}
		return t, true, nil
	}

	// DateTime with Z: YYYYMMDDTHHMMSSZ
	if strings.HasSuffix(s, "Z") {
		t, err := time.Parse("20060102T150405Z", s)
		if err != nil {
			return time.Time{}, false, err
		}
		return t, false, nil
	}

	// DateTime without Z: YYYYMMDDTHHMMSS (treat as UTC)
	if len(s) == 15 && strings.Contains(s, "T") {
		t, err := time.Parse("20060102T150405", s)
		if err != nil {
			return time.Time{}, false, err
		}
		return t.UTC(), false, nil
	}

	return time.Time{}, false, fmt.Errorf("unsupported datetime format: %s", s)
}

func mapRRule(rrule string) string {
	for _, part := range strings.Split(rrule, ";") {
		if strings.HasPrefix(part, "FREQ=") {
			freq := strings.TrimPrefix(part, "FREQ=")
			switch freq {
			case "DAILY":
				return "daily"
			case "WEEKLY":
				return "weekly"
			case "MONTHLY":
				return "monthly"
			case "YEARLY":
				return "yearly"
			}
		}
	}
	return ""
}
