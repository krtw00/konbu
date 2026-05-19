package service

import (
	"encoding/json"
	"testing"
)

func TestResolveNotificationSettings_Defaults(t *testing.T) {
	// Empty user_settings -> opt-out (enabled=false), fall back to account email.
	s, to := resolveNotificationSettings(nil, "user@example.com")
	if s.Enabled {
		t.Fatalf("expected enabled=false by default, got true")
	}
	if s.EventLeadMinutes != defaultEventLeadMinutes {
		t.Fatalf("expected lead=%d, got %d", defaultEventLeadMinutes, s.EventLeadMinutes)
	}
	if s.TodoDueTime != defaultTodoDueTime {
		t.Fatalf("expected todo_due_time=%q, got %q", defaultTodoDueTime, s.TodoDueTime)
	}
	if s.Timezone != defaultTimezone {
		t.Fatalf("expected timezone=%q, got %q", defaultTimezone, s.Timezone)
	}
	if to != "user@example.com" {
		t.Fatalf("expected to fall back to account email, got %q", to)
	}
}

func TestResolveNotificationSettings_Override(t *testing.T) {
	raw := json.RawMessage(`{"notifications":{"enabled":true,"email":"alerts@example.com","event_lead_minutes":15,"todo_due_time":"07:30","timezone":"UTC"}}`)
	s, to := resolveNotificationSettings(raw, "user@example.com")
	if !s.Enabled {
		t.Fatalf("expected enabled=true")
	}
	if s.EventLeadMinutes != 15 {
		t.Fatalf("expected lead=15, got %d", s.EventLeadMinutes)
	}
	if s.TodoDueTime != "07:30" {
		t.Fatalf("expected todo_due_time=07:30, got %q", s.TodoDueTime)
	}
	if s.Timezone != "UTC" {
		t.Fatalf("expected timezone=UTC, got %q", s.Timezone)
	}
	if to != "alerts@example.com" {
		t.Fatalf("expected configured email, got %q", to)
	}
}

func TestResolveNotificationSettings_PartialKeepsDefaults(t *testing.T) {
	// Missing event_lead_minutes / todo_due_time / timezone keep their defaults.
	raw := json.RawMessage(`{"notifications":{"enabled":true}}`)
	s, to := resolveNotificationSettings(raw, "user@example.com")
	if !s.Enabled {
		t.Fatalf("expected enabled=true")
	}
	if s.EventLeadMinutes != defaultEventLeadMinutes {
		t.Fatalf("expected default lead, got %d", s.EventLeadMinutes)
	}
	if s.TodoDueTime != defaultTodoDueTime {
		t.Fatalf("expected default todo_due_time, got %q", s.TodoDueTime)
	}
	if s.Timezone != defaultTimezone {
		t.Fatalf("expected default timezone, got %q", s.Timezone)
	}
	if to != "user@example.com" {
		t.Fatalf("expected fall back, got %q", to)
	}
}

func TestParseHHMM(t *testing.T) {
	tests := []struct {
		in    string
		h, m  int
		valid bool
	}{
		{"09:00", 9, 0, true},
		{"00:00", 0, 0, true},
		{"23:59", 23, 59, true},
		{"24:00", 0, 0, false},
		{"12:60", 0, 0, false},
		{"-1:00", 0, 0, false},
		{"", 0, 0, false},
		{"nope", 0, 0, false},
	}
	for _, tt := range tests {
		h, m, ok := parseHHMM(tt.in)
		if ok != tt.valid {
			t.Fatalf("parseHHMM(%q) ok=%v, want %v", tt.in, ok, tt.valid)
		}
		if ok && (h != tt.h || m != tt.m) {
			t.Fatalf("parseHHMM(%q) = (%d,%d), want (%d,%d)", tt.in, h, m, tt.h, tt.m)
		}
	}
}

func TestMimeEncodeHeader(t *testing.T) {
	if got := mimeEncodeHeader("Hello"); got != "Hello" {
		t.Fatalf("ascii unchanged, got %q", got)
	}
	got := mimeEncodeHeader("こんにちは")
	want := "=?UTF-8?B?44GT44KT44Gr44Gh44Gv?="
	if got != want {
		t.Fatalf("non-ascii encoded: got %q, want %q", got, want)
	}
}
