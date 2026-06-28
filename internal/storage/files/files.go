package files

import (
	"context"
	"crypto/sha1"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Daniel1212649/LinksHelperBot/internal/lib/e"
	"github.com/Daniel1212649/LinksHelperBot/internal/storage"
)

const defaultPerm = 0774

type Storage struct {
	basePath string
}

func New(basePath string) Storage {
	if err := os.MkdirAll(basePath, defaultPerm); err != nil {
		panic(fmt.Sprintf("can't create base storage path: %v", err))
	}
	return Storage{basePath: basePath}
}

func (s Storage) Save(_ context.Context, user storage.User, rawURL string) (page *storage.Page, err error) {
	return s.SaveWithDetails(context.Background(), user, rawURL, "", "", "", nil)
}

func (s Storage) SaveWithDetails(_ context.Context, user storage.User, rawURL string, title string, note string, groupName string, remindAt *time.Time) (page *storage.Page, err error) {
	defer func() { err = e.WrapIfErr("can't save page", err) }()

	normalizedURL, err := storage.NormalizeURL(rawURL)
	if err != nil {
		return nil, err
	}

	page = &storage.Page{
		URL:           normalizedURL,
		NormalizedURL: normalizedURL,
		Title:         title,
		Note:          note,
		GroupName:     storage.NormalizeGroupName(groupName),
		Status:        storage.StatusUnread,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		RemindAt:      remindAt,
	}

	if exists, err := s.isExists(user, page); err != nil {
		return nil, err
	} else if exists {
		return nil, storage.ErrPageExists
	}

	userDir := filepath.Join(s.basePath, userKey(user))
	if err := os.MkdirAll(userDir, defaultPerm); err != nil {
		return nil, err
	}

	path := filepath.Join(userDir, fileName(user, page.NormalizedURL))
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return nil, err
	}

	return page, nil
}

func (s Storage) PickRandom(_ context.Context, user storage.User) (page *storage.Page, err error) {
	defer func() { err = e.WrapIfErr("can't pick random page", err) }()

	path := filepath.Join(s.basePath, userKey(user))
	files, err := os.ReadDir(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, storage.ErrNoSavedPages
	}
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	file := files[rand.Intn(len(files))]
	return s.decodePage(filepath.Join(path, file.Name()))
}

func (s Storage) MarkRead(_ context.Context, user storage.User, id int64) error {
	pages, err := s.List(context.Background(), user, 1000)
	if err != nil {
		return err
	}
	for _, page := range pages {
		if page.ID == id {
			page.Status = storage.StatusRead
			now := time.Now()
			page.ReadAt = &now
			page.UpdatedAt = now
			return s.rewrite(user, &page)
		}
	}
	return nil
}

func (s Storage) Remove(_ context.Context, user storage.User, id int64) error {
	pages, err := s.List(context.Background(), user, 1000)
	if err != nil {
		return err
	}
	for _, page := range pages {
		if page.ID == id {
			path := filepath.Join(s.basePath, userKey(user), fileName(user, page.NormalizedURL))
			return os.Remove(path)
		}
	}
	return nil
}

func (s Storage) List(_ context.Context, user storage.User, limit int) ([]storage.Page, error) {
	if limit <= 0 {
		limit = 10
	}

	path := filepath.Join(s.basePath, userKey(user))
	files, err := os.ReadDir(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	pages := make([]storage.Page, 0, len(files))
	for _, file := range files {
		page, err := s.decodePage(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}
		if page.Status == "" {
			page.Status = storage.StatusUnread
		}
		pages = append(pages, *page)
		if len(pages) >= limit {
			break
		}
	}

	return pages, nil
}

func (s Storage) ListByGroup(ctx context.Context, user storage.User, groupName string, limit int) ([]storage.Page, error) {
	pages, err := s.List(ctx, user, 1000)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}

	groupName = storage.NormalizeGroupName(groupName)
	res := make([]storage.Page, 0)
	for _, page := range pages {
		if strings.EqualFold(page.GroupName, groupName) {
			res = append(res, page)
		}
		if len(res) >= limit {
			break
		}
	}
	return res, nil
}

func (s Storage) Search(ctx context.Context, user storage.User, query string, limit int) ([]storage.Page, error) {
	pages, err := s.List(ctx, user, 1000)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}

	res := make([]storage.Page, 0)
	for _, page := range pages {
		if stringsContains(page.URL, query) || stringsContains(page.Title, query) || stringsContains(page.Note, query) || stringsContains(page.GroupName, query) {
			res = append(res, page)
		}
		if len(res) >= limit {
			break
		}
	}
	return res, nil
}

func (s Storage) Stats(ctx context.Context, user storage.User) (storage.Stats, error) {
	pages, err := s.List(ctx, user, 1000)
	if err != nil {
		return storage.Stats{}, err
	}

	stats := storage.Stats{Total: int64(len(pages))}
	for _, page := range pages {
		switch page.Status {
		case storage.StatusRead:
			stats.Read++
		default:
			stats.Unread++
		}
	}

	return stats, nil
}

func (s Storage) Groups(ctx context.Context, user storage.User) ([]storage.Group, error) {
	pages, err := s.List(ctx, user, 1000)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	names := make([]string, 0)
	for _, page := range pages {
		groupName := storage.NormalizeGroupName(page.GroupName)
		if groupName == "" {
			continue
		}
		if _, ok := counts[strings.ToLower(groupName)]; !ok {
			names = append(names, groupName)
		}
		counts[strings.ToLower(groupName)]++
	}

	groups := make([]storage.Group, 0, len(names))
	for _, name := range names {
		groups = append(groups, storage.Group{Name: name, Count: counts[strings.ToLower(name)]})
	}
	return groups, nil
}

func (s Storage) SetNote(ctx context.Context, user storage.User, id int64, note string) error {
	pages, err := s.List(ctx, user, 1000)
	if err != nil {
		return err
	}
	for _, page := range pages {
		if page.ID == id {
			page.Note = note
			page.UpdatedAt = time.Now()
			return s.rewrite(user, &page)
		}
	}
	return nil
}

func (s Storage) SetGroup(ctx context.Context, user storage.User, id int64, groupName string) error {
	pages, err := s.List(ctx, user, 1000)
	if err != nil {
		return err
	}
	for _, page := range pages {
		if page.ID == id {
			page.GroupName = storage.NormalizeGroupName(groupName)
			page.UpdatedAt = time.Now()
			return s.rewrite(user, &page)
		}
	}
	return nil
}

func (s Storage) SetReminder(ctx context.Context, user storage.User, id int64, remindAt time.Time) error {
	pages, err := s.List(ctx, user, 1000)
	if err != nil {
		return err
	}
	for _, page := range pages {
		if page.ID == id {
			page.RemindAt = &remindAt
			page.RemindedAt = nil
			page.UpdatedAt = time.Now()
			return s.rewrite(user, &page)
		}
	}
	return nil
}

func (s Storage) DueReminders(context.Context, time.Time, int) ([]storage.Reminder, error) {
	return nil, nil
}

func (s Storage) MarkReminded(context.Context, int64) error {
	return nil
}

func (s Storage) GetLocale(_ context.Context, user storage.User) (string, error) {
	return storage.NormalizeLocale(user.Locale), nil
}

func (s Storage) SetLocale(_ context.Context, _ storage.User, _ string) error {
	return nil
}

func (s Storage) isExists(user storage.User, page *storage.Page) (bool, error) {
	path := filepath.Join(s.basePath, userKey(user), fileName(user, page.NormalizedURL))
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

func (s Storage) rewrite(user storage.User, page *storage.Page) error {
	path := filepath.Join(s.basePath, userKey(user), fileName(user, page.NormalizedURL))
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	return gob.NewEncoder(file).Encode(page)
}

func (s Storage) decodePage(filePath string) (*storage.Page, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, e.Wrap("can't decode page", err)
	}
	defer func() { _ = file.Close() }()

	var page storage.Page
	if err := gob.NewDecoder(file).Decode(&page); err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	return &page, nil
}

func fileName(user storage.User, normalizedURL string) string {
	hash := sha1.Sum([]byte(userKey(user) + normalizedURL))
	return fmt.Sprintf("%x", hash)
}

func userKey(user storage.User) string {
	if user.TelegramID != 0 {
		return strconv.FormatInt(user.TelegramID, 10)
	}
	if user.ChatID != 0 {
		return strconv.FormatInt(user.ChatID, 10)
	}
	if user.Username != "" {
		return user.Username
	}
	return "unknown"
}

func stringsContains(value string, query string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(query))
}
