package telegram

import "testing"

func TestParseCallbackID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    string
		prefix  string
		want    int64
		wantErr bool
	}{
		{name: "read", data: "read:42", prefix: "read:", want: 42},
		{name: "delete", data: "del:7", prefix: "del:", want: 7},
		{name: "note", data: "note:8", prefix: "note:", want: 8},
		{name: "remind", data: "remind:9", prefix: "remind:", want: 9},
		{name: "group", data: "group:10", prefix: "group:", want: 10},
		{name: "invalid", data: "read:abc", prefix: "read:", wantErr: true},
		{name: "zero", data: "read:0", prefix: "read:", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseCallbackID(tt.data, tt.prefix)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMainMenuKeyboard(t *testing.T) {
	t.Parallel()

	kb := mainMenuKeyboard("ru")
	if len(kb.InlineKeyboard) != 7 {
		t.Fatalf("expected 7 rows, got %d", len(kb.InlineKeyboard))
	}
	if kb.InlineKeyboard[1][0].CallbackData != cbCmdSave {
		t.Fatalf("unexpected save callback: %q", kb.InlineKeyboard[1][0].CallbackData)
	}
	if kb.InlineKeyboard[2][0].CallbackData != cbCmdNote {
		t.Fatalf("unexpected note callback: %q", kb.InlineKeyboard[2][0].CallbackData)
	}
	if kb.InlineKeyboard[3][0].CallbackData != cbCmdGroups {
		t.Fatalf("unexpected groups callback: %q", kb.InlineKeyboard[3][0].CallbackData)
	}
	if kb.InlineKeyboard[5][1].CallbackData != cbCmdDelete {
		t.Fatalf("unexpected delete callback: %q", kb.InlineKeyboard[5][1].CallbackData)
	}
	if kb.InlineKeyboard[6][0].CallbackData != cbCmdLang {
		t.Fatalf("unexpected language callback: %q", kb.InlineKeyboard[6][0].CallbackData)
	}
}

func TestEnglishMainMenuKeyboard(t *testing.T) {
	t.Parallel()

	kb := mainMenuKeyboard("en")
	if got := kb.InlineKeyboard[0][0].Text; got != "👋 Start" {
		t.Fatalf("unexpected start button text: %q", got)
	}
}

func TestLinkActionKeyboard(t *testing.T) {
	t.Parallel()

	kb := linkActionKeyboard("ru", 15)
	if len(kb.InlineKeyboard) != 9 {
		t.Fatalf("expected 9 rows, got %d", len(kb.InlineKeyboard))
	}
	if kb.InlineKeyboard[0][0].CallbackData != "read:15" {
		t.Fatalf("unexpected read callback: %q", kb.InlineKeyboard[0][0].CallbackData)
	}
	if kb.InlineKeyboard[0][1].CallbackData != "del:15" {
		t.Fatalf("unexpected delete callback: %q", kb.InlineKeyboard[0][1].CallbackData)
	}
	if kb.InlineKeyboard[1][0].CallbackData != "note:15" {
		t.Fatalf("unexpected note callback: %q", kb.InlineKeyboard[1][0].CallbackData)
	}
	if kb.InlineKeyboard[1][1].CallbackData != "remind:15" {
		t.Fatalf("unexpected remind callback: %q", kb.InlineKeyboard[1][1].CallbackData)
	}
	if kb.InlineKeyboard[1][2].CallbackData != "group:15" {
		t.Fatalf("unexpected group callback: %q", kb.InlineKeyboard[1][2].CallbackData)
	}
}
