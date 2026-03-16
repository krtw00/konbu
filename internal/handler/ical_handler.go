package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/service"
)

type ICalHandler struct {
	authSvc  *service.AuthService
	eventSvc *service.EventService
}

func NewICalHandler(authSvc *service.AuthService, eventSvc *service.EventService) *ICalHandler {
	return &ICalHandler{authSvc: authSvc, eventSvc: eventSvc}
}

func (h *ICalHandler) HandleCalendarICS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	user, err := h.authSvc.AuthenticateByAPIKey(r.Context(), token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	events, err := h.eventSvc.ListAllEvents(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	ical := buildICalendar(events)

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "inline; filename=\"calendar.ics\"")
	w.Write([]byte(ical))
}

func buildICalendar(events []model.CalendarEvent) string {
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//konbu//konbu//EN\r\n")
	b.WriteString("CALSCALE:GREGORIAN\r\n")
	b.WriteString("METHOD:PUBLISH\r\n")

	for _, e := range events {
		b.WriteString("BEGIN:VEVENT\r\n")
		b.WriteString(fmt.Sprintf("UID:%s@konbu\r\n", e.ID.String()))
		if e.AllDay {
			b.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", e.StartAt.UTC().Format("20060102")))
			if e.EndAt != nil {
				b.WriteString(fmt.Sprintf("DTEND;VALUE=DATE:%s\r\n", e.EndAt.UTC().Format("20060102")))
			}
		} else {
			b.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatICalTime(e.StartAt)))
			if e.EndAt != nil {
				b.WriteString(fmt.Sprintf("DTEND:%s\r\n", formatICalTime(*e.EndAt)))
			}
		}
		b.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICalText(e.Title)))
		if e.Description != "" {
			b.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICalText(e.Description)))
		}
		if e.RecurrenceRule != nil && *e.RecurrenceRule != "" {
			b.WriteString(fmt.Sprintf("RRULE:%s\r\n", *e.RecurrenceRule))
		}
		b.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", formatICalTime(e.CreatedAt)))
		b.WriteString("END:VEVENT\r\n")
	}

	b.WriteString("END:VCALENDAR\r\n")
	return b.String()
}

func formatICalTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

func escapeICalText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
