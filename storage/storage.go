package storage

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"
)

var (
	ErrNoSavedPages = errors.New("no saved pages")
	ErrPageExists   = errors.New("page already exists")
)

const (
	StatusUnread  = "unread"
	StatusRead    = "read"
	StatusDeleted = "deleted"
)

const (
	LocaleRU = "ru"
	LocaleEN = "en"
)

type Storage interface {
	Save(ctx context.Context, user User, rawURL string) (*Page, error)
	PickRandom(ctx context.Context, user User) (*Page, error)
	MarkRead(ctx context.Context, user User, id int64) error
	Remove(ctx context.Context, user User, id int64) error
	List(ctx context.Context, user User, limit int) ([]Page, error)
	Search(ctx context.Context, user User, query string, limit int) ([]Page, error)
	Stats(ctx context.Context, user User) (Stats, error)
	GetLocale(ctx context.Context, user User) (string, error)
	SetLocale(ctx context.Context, user User, locale string) error
}

type User struct {
	TelegramID int64
	ChatID     int64
	Username   string
	Locale     string
}

type Page struct {
	ID            int64
	URL           string
	NormalizedURL string
	Title         string
	Description   string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ReadAt        *time.Time
}

type Stats struct {
	Total  int64
	Unread int64
	Read   int64
}

func NormalizeLocale(locale string) string {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case LocaleEN, "en-us", "en-gb":
		return LocaleEN
	default:
		return LocaleRU
	}
}

func NormalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" {
		parsed, err = url.Parse("https://" + raw)
		if err != nil {
			return "", err
		}
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("only http and https links are supported")
	}
	if parsed.Host == "" {
		return "", errors.New("link host is required")
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""

	query := parsed.Query()
	for key := range query {
		lowerKey := strings.ToLower(key)
		if strings.HasPrefix(lowerKey, "utm_") || lowerKey == "fbclid" || lowerKey == "gclid" {
			query.Del(key)
		}
	}
	parsed.RawQuery = query.Encode()

	if parsed.Path == "/" {
		parsed.Path = ""
	}

	return parsed.String(), nil
}
