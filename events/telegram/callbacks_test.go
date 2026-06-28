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

	kb := mainMenuKeyboard()
	if len(kb.InlineKeyboard) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(kb.InlineKeyboard))
	}
	if kb.InlineKeyboard[1][0].CallbackData != cbCmdSave {
		t.Fatalf("unexpected save callback: %q", kb.InlineKeyboard[1][0].CallbackData)
	}
	if kb.InlineKeyboard[3][1].CallbackData != cbCmdDelete {
		t.Fatalf("unexpected delete callback: %q", kb.InlineKeyboard[3][1].CallbackData)
	}
}

func TestLinkActionKeyboard(t *testing.T) {
	t.Parallel()

	kb := linkActionKeyboard(15)
	if len(kb.InlineKeyboard) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(kb.InlineKeyboard))
	}
	if kb.InlineKeyboard[0][0].CallbackData != "read:15" {
		t.Fatalf("unexpected read callback: %q", kb.InlineKeyboard[0][0].CallbackData)
	}
	if kb.InlineKeyboard[0][1].CallbackData != "del:15" {
		t.Fatalf("unexpected delete callback: %q", kb.InlineKeyboard[0][1].CallbackData)
	}
}
