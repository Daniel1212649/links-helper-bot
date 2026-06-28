package storage

import "testing"

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "adds scheme",
			in:   "Example.com/articles?id=10",
			want: "https://example.com/articles?id=10",
		},
		{
			name: "normalizes host and removes fragment",
			in:   "HTTPS://Example.COM/path#section",
			want: "https://example.com/path",
		},
		{
			name: "removes tracking query params",
			in:   "https://example.com/?utm_source=newsletter&id=42&fbclid=abc",
			want: "https://example.com?id=42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeURL(tt.in)
			if err != nil {
				t.Fatalf("NormalizeURL returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeURLRejectsUnsupportedScheme(t *testing.T) {
	if _, err := NormalizeURL("ftp://example.com/file"); err == nil {
		t.Fatal("expected unsupported scheme error")
	}
}

func TestNormalizeGroupName(t *testing.T) {
	if got := NormalizeGroupName("  Go articles  "); got != "Go articles" {
		t.Fatalf("NormalizeGroupName() = %q, want %q", got, "Go articles")
	}
}
