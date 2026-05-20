package service

import (
	"strings"
	"testing"
)

func TestParseICalExtractsUID(t *testing.T) {
	ics := "BEGIN:VCALENDAR\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:abc-123@example.com\r\n" +
		"SUMMARY:Meeting\r\n" +
		"DTSTART:20260101T090000Z\r\n" +
		"END:VEVENT\r\n" +
		"BEGIN:VEVENT\r\n" +
		"SUMMARY:No UID Event\r\n" +
		"DTSTART:20260102T090000Z\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n"

	events, err := parseICal(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("parseICal returned error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].uid != "abc-123@example.com" {
		t.Errorf("event 0 uid = %q, want %q", events[0].uid, "abc-123@example.com")
	}
	if events[0].summary != "Meeting" {
		t.Errorf("event 0 summary = %q, want %q", events[0].summary, "Meeting")
	}
	if events[1].uid != "" {
		t.Errorf("event 1 uid = %q, want empty", events[1].uid)
	}
}

func TestValidateICalURL(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https", "https://calendar.google.com/calendar/ical/x/private/basic.ics", false},
		{"http rejected", "http://example.com/cal.ics", true},
		{"empty rejected", "", true},
		{"no host rejected", "https://", true},
		{"non-url scheme rejected", "ftp://example.com/cal.ics", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateICalURL(c.url)
			if c.wantErr && err == nil {
				t.Errorf("validateICalURL(%q) = nil, want error", c.url)
			}
			if !c.wantErr && err != nil {
				t.Errorf("validateICalURL(%q) = %v, want nil", c.url, err)
			}
		})
	}
}
