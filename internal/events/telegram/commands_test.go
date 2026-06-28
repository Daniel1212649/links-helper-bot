package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseSaveArgs(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 28, 13, 0, 0, 0, moscowLocation)

	tests := []struct {
		name        string
		in          string
		wantURL     string
		wantNote    string
		wantGroup   string
		wantRemind  string
		wantMatched bool
		wantErr     bool
	}{
		{
			name:        "url only",
			in:          "https://example.com",
			wantURL:     "https://example.com",
			wantMatched: true,
		},
		{
			name:        "url with note",
			in:          "https://example.com useful article",
			wantURL:     "https://example.com",
			wantNote:    "useful article",
			wantMatched: true,
		},
		{
			name:        "url with group",
			in:          "https://example.com useful article --group Go",
			wantURL:     "https://example.com",
			wantNote:    "useful article",
			wantGroup:   "Go",
			wantMatched: true,
		},
		{
			name:        "url with group and reminder",
			in:          "https://example.com useful article --group Go --remind 2026-07-01 09:30",
			wantURL:     "https://example.com",
			wantNote:    "useful article",
			wantGroup:   "Go",
			wantRemind:  "2026-07-01 09:30",
			wantMatched: true,
		},
		{
			name:        "url with reminder before group",
			in:          "https://example.com useful article --remind 2026-07-01 09:30 --group Go",
			wantURL:     "https://example.com",
			wantNote:    "useful article",
			wantGroup:   "Go",
			wantRemind:  "2026-07-01 09:30",
			wantMatched: true,
		},
		{
			name:        "url with note and reminder",
			in:          "https://example.com useful article --remind 2026-07-01 09:30",
			wantURL:     "https://example.com",
			wantNote:    "useful article",
			wantRemind:  "2026-07-01 09:30",
			wantMatched: true,
		},
		{
			name:        "url with em dash reminder marker",
			in:          "https://example.com useful article —remind 2026-07-01 09:30",
			wantURL:     "https://example.com",
			wantNote:    "useful article",
			wantRemind:  "2026-07-01 09:30",
			wantMatched: true,
		},
		{
			name:        "url with reminder only",
			in:          "https://example.com --remind 2026-07-01",
			wantURL:     "https://example.com",
			wantRemind:  "2026-07-01 09:00",
			wantMatched: true,
		},
		{
			name: "not a link",
			in:   "hello world",
		},
		{
			name:        "invalid reminder date",
			in:          "https://example.com note --remind tomorrow",
			wantMatched: true,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotURL, gotNote, gotGroup, gotRemind, gotMatched, err := parseSaveArgs(tt.in, now)
			if gotMatched != tt.wantMatched {
				t.Fatalf("matched = %v, want %v", gotMatched, tt.wantMatched)
			}
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotURL != tt.wantURL {
				t.Fatalf("url = %q, want %q", gotURL, tt.wantURL)
			}
			if gotNote != tt.wantNote {
				t.Fatalf("note = %q, want %q", gotNote, tt.wantNote)
			}
			if gotGroup != tt.wantGroup {
				t.Fatalf("group = %q, want %q", gotGroup, tt.wantGroup)
			}
			if gotRemind == nil && tt.wantRemind != "" {
				t.Fatalf("expected reminder %q", tt.wantRemind)
			}
			if gotRemind != nil && formatReminderTime(*gotRemind) != tt.wantRemind {
				t.Fatalf("reminder = %q, want %q", formatReminderTime(*gotRemind), tt.wantRemind)
			}
		})
	}
}

func TestFetchPageTitle(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><head><title> Go downloads &amp; releases </title></head><body></body></html>`))
	}))
	defer server.Close()

	got, err := fetchPageTitle(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("fetchPageTitle returned error: %v", err)
	}
	if got != "Go downloads & releases" {
		t.Fatalf("title = %q, want %q", got, "Go downloads & releases")
	}
}

func TestParseReminderTimeUsesMoscowTimezone(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 28, 13, 0, 0, 0, moscowLocation)
	got, err := parseReminderTime("2026-07-01 09:30", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Location().String() != moscowTimezone {
		t.Fatalf("location = %q, want %q", got.Location().String(), moscowTimezone)
	}
	if got.UTC().Format(time.RFC3339) != "2026-07-01T06:30:00Z" {
		t.Fatalf("utc time = %q, want 2026-07-01T06:30:00Z", got.UTC().Format(time.RFC3339))
	}
}
