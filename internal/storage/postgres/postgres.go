package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Daniel1212649/LinksHelperBot/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505"

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}

func (s *Storage) Save(ctx context.Context, user storage.User, rawURL string) (*storage.Page, error) {
	return s.SaveWithDetails(ctx, user, rawURL, "", "", "", nil)
}

func (s *Storage) SaveWithDetails(ctx context.Context, user storage.User, rawURL string, title string, note string, groupName string, remindAt *time.Time) (*storage.Page, error) {
	normalizedURL, err := storage.NormalizeURL(rawURL)
	if err != nil {
		return nil, err
	}

	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return nil, err
	}
	groupName = storage.NormalizeGroupName(groupName)

	const query = `
		insert into links (user_id, url, normalized_url, title, note, group_name, remind_at, status)
		values ($1, $2, $3, nullif($4, ''), $5, $6, $7, $8)
		on conflict (user_id, normalized_url) do update
		set url = excluded.url,
			title = coalesce(excluded.title, links.title),
			note = excluded.note,
			group_name = excluded.group_name,
			remind_at = excluded.remind_at,
			reminded_at = null,
			status = excluded.status,
			read_at = null,
			updated_at = now()
		where links.status = $9
		returning id, url, normalized_url, coalesce(title, ''), coalesce(description, ''), note, group_name, status, created_at, updated_at, read_at, remind_at, reminded_at`

	var page storage.Page
	err = s.pool.QueryRow(ctx, query, userID, normalizedURL, normalizedURL, title, note, groupName, remindAt, storage.StatusUnread, storage.StatusDeleted).Scan(
		&page.ID,
		&page.URL,
		&page.NormalizedURL,
		&page.Title,
		&page.Description,
		&page.Note,
		&page.GroupName,
		&page.Status,
		&page.CreatedAt,
		&page.UpdatedAt,
		&page.ReadAt,
		&page.RemindAt,
		&page.RemindedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, storage.ErrPageExists
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return nil, storage.ErrPageExists
		}
		return nil, err
	}

	return &page, nil
}

func (s *Storage) PickRandom(ctx context.Context, user storage.User) (*storage.Page, error) {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return nil, err
	}

	const query = `
		select id, url, normalized_url, coalesce(title, ''), coalesce(description, ''), note, group_name, status, created_at, updated_at, read_at, remind_at, reminded_at
		from links
		where user_id = $1 and status = $2
		order by random()
		limit 1`

	var page storage.Page
	err = s.pool.QueryRow(ctx, query, userID, storage.StatusUnread).Scan(
		&page.ID,
		&page.URL,
		&page.NormalizedURL,
		&page.Title,
		&page.Description,
		&page.Note,
		&page.GroupName,
		&page.Status,
		&page.CreatedAt,
		&page.UpdatedAt,
		&page.ReadAt,
		&page.RemindAt,
		&page.RemindedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, storage.ErrNoSavedPages
	}
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (s *Storage) MarkRead(ctx context.Context, user storage.User, id int64) error {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return err
	}

	const query = `
		update links
		set status = $1, read_at = now(), updated_at = now()
		where id = $2 and user_id = $3 and status <> $4`

	_, err = s.pool.Exec(ctx, query, storage.StatusRead, id, userID, storage.StatusDeleted)
	return err
}

func (s *Storage) Remove(ctx context.Context, user storage.User, id int64) error {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return err
	}

	const query = `
		update links
		set status = $1, updated_at = now()
		where id = $2 and user_id = $3`

	_, err = s.pool.Exec(ctx, query, storage.StatusDeleted, id, userID)
	return err
}

func (s *Storage) List(ctx context.Context, user storage.User, limit int) ([]storage.Page, error) {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	const query = `
		select id, url, normalized_url, coalesce(title, ''), coalesce(description, ''), note, group_name, status, created_at, updated_at, read_at, remind_at, reminded_at
		from links
		where user_id = $1 and status <> $2
		order by created_at desc
		limit $3`

	rows, err := s.pool.Query(ctx, query, userID, storage.StatusDeleted, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPages(rows)
}

func (s *Storage) ListByGroup(ctx context.Context, user storage.User, groupName string, limit int) ([]storage.Page, error) {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	groupName = storage.NormalizeGroupName(groupName)

	const query = `
		select id, url, normalized_url, coalesce(title, ''), coalesce(description, ''), note, group_name, status, created_at, updated_at, read_at, remind_at, reminded_at
		from links
		where user_id = $1
			and status <> $2
			and lower(group_name) = lower($3)
		order by created_at desc
		limit $4`

	rows, err := s.pool.Query(ctx, query, userID, storage.StatusDeleted, groupName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPages(rows)
}

func (s *Storage) Search(ctx context.Context, user storage.User, queryText string, limit int) ([]storage.Page, error) {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	const query = `
		select id, url, normalized_url, coalesce(title, ''), coalesce(description, ''), note, group_name, status, created_at, updated_at, read_at, remind_at, reminded_at
		from links
		where user_id = $1
			and status <> $2
			and (url ilike '%' || $3 || '%' or coalesce(title, '') ilike '%' || $3 || '%' or note ilike '%' || $3 || '%' or group_name ilike '%' || $3 || '%')
		order by created_at desc
		limit $4`

	rows, err := s.pool.Query(ctx, query, userID, storage.StatusDeleted, queryText, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPages(rows)
}

func (s *Storage) Stats(ctx context.Context, user storage.User) (storage.Stats, error) {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return storage.Stats{}, err
	}

	const query = `
		select
			count(*) filter (where status <> $2) as total,
			count(*) filter (where status = $3) as unread,
			count(*) filter (where status = $4) as read
		from links
		where user_id = $1`

	var stats storage.Stats
	err = s.pool.QueryRow(ctx, query, userID, storage.StatusDeleted, storage.StatusUnread, storage.StatusRead).Scan(
		&stats.Total,
		&stats.Unread,
		&stats.Read,
	)
	return stats, err
}

func (s *Storage) Groups(ctx context.Context, user storage.User) ([]storage.Group, error) {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return nil, err
	}

	const query = `
		select group_name, count(*)
		from links
		where user_id = $1
			and status <> $2
			and group_name <> ''
		group by group_name
		order by lower(group_name)`

	rows, err := s.pool.Query(ctx, query, userID, storage.StatusDeleted)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]storage.Group, 0)
	for rows.Next() {
		var group storage.Group
		if err := rows.Scan(&group.Name, &group.Count); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

func (s *Storage) GetLocale(ctx context.Context, user storage.User) (string, error) {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return storage.LocaleRU, err
	}

	const query = `select locale from users where id = $1`

	var locale string
	if err := s.pool.QueryRow(ctx, query, userID).Scan(&locale); err != nil {
		return storage.LocaleRU, err
	}
	return storage.NormalizeLocale(locale), nil
}

func (s *Storage) SetLocale(ctx context.Context, user storage.User, locale string) error {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return err
	}

	const query = `
		update users
		set locale = $1, updated_at = now()
		where id = $2`

	_, err = s.pool.Exec(ctx, query, storage.NormalizeLocale(locale), userID)
	return err
}

func (s *Storage) SetNote(ctx context.Context, user storage.User, id int64, note string) error {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return err
	}

	const query = `
		update links
		set note = $1, updated_at = now()
		where id = $2 and user_id = $3 and status <> $4`

	_, err = s.pool.Exec(ctx, query, note, id, userID, storage.StatusDeleted)
	return err
}

func (s *Storage) SetGroup(ctx context.Context, user storage.User, id int64, groupName string) error {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return err
	}

	const query = `
		update links
		set group_name = $1, updated_at = now()
		where id = $2 and user_id = $3 and status <> $4`

	_, err = s.pool.Exec(ctx, query, storage.NormalizeGroupName(groupName), id, userID, storage.StatusDeleted)
	return err
}

func (s *Storage) SetReminder(ctx context.Context, user storage.User, id int64, remindAt time.Time) error {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return err
	}

	const query = `
		update links
		set remind_at = $1, reminded_at = null, updated_at = now()
		where id = $2 and user_id = $3 and status <> $4`

	_, err = s.pool.Exec(ctx, query, remindAt, id, userID, storage.StatusDeleted)
	return err
}

func (s *Storage) DueReminders(ctx context.Context, now time.Time, limit int) ([]storage.Reminder, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	const query = `
		select
			l.id, l.url, l.normalized_url, coalesce(l.title, ''), coalesce(l.description, ''), l.note, l.group_name,
			l.status, l.created_at, l.updated_at, l.read_at, l.remind_at, l.reminded_at,
			coalesce(u.telegram_user_id, 0), u.chat_id, coalesce(u.username, ''), u.locale
		from links l
		join users u on u.id = l.user_id
		where l.status <> $1
			and l.remind_at is not null
			and l.reminded_at is null
			and l.remind_at <= $2
		order by l.remind_at asc
		limit $3`

	rows, err := s.pool.Query(ctx, query, storage.StatusDeleted, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reminders := make([]storage.Reminder, 0)
	for rows.Next() {
		var reminder storage.Reminder
		if err := rows.Scan(
			&reminder.Page.ID,
			&reminder.Page.URL,
			&reminder.Page.NormalizedURL,
			&reminder.Page.Title,
			&reminder.Page.Description,
			&reminder.Page.Note,
			&reminder.Page.GroupName,
			&reminder.Page.Status,
			&reminder.Page.CreatedAt,
			&reminder.Page.UpdatedAt,
			&reminder.Page.ReadAt,
			&reminder.Page.RemindAt,
			&reminder.Page.RemindedAt,
			&reminder.User.TelegramID,
			&reminder.User.ChatID,
			&reminder.User.Username,
			&reminder.User.Locale,
		); err != nil {
			return nil, err
		}
		reminders = append(reminders, reminder)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return reminders, nil
}

func (s *Storage) MarkReminded(ctx context.Context, id int64) error {
	const query = `
		update links
		set reminded_at = now(), updated_at = now()
		where id = $1`

	_, err := s.pool.Exec(ctx, query, id)
	return err
}

func (s *Storage) ensureUser(ctx context.Context, user storage.User) (int64, error) {
	const query = `
		insert into users (telegram_user_id, username, chat_id, locale)
		values (nullif($1, 0), nullif($2, ''), $3, $4)
		on conflict (chat_id) do update
		set telegram_user_id = coalesce(excluded.telegram_user_id, users.telegram_user_id),
			username = coalesce(excluded.username, users.username),
			updated_at = now()
		returning id`

	var id int64
	err := s.pool.QueryRow(ctx, query, user.TelegramID, user.Username, user.ChatID, storage.NormalizeLocale(user.Locale)).Scan(&id)
	return id, err
}

func scanPages(rows pgx.Rows) ([]storage.Page, error) {
	pages := make([]storage.Page, 0)
	for rows.Next() {
		var page storage.Page
		if err := rows.Scan(
			&page.ID,
			&page.URL,
			&page.NormalizedURL,
			&page.Title,
			&page.Description,
			&page.Note,
			&page.GroupName,
			&page.Status,
			&page.CreatedAt,
			&page.UpdatedAt,
			&page.ReadAt,
			&page.RemindAt,
			&page.RemindedAt,
		); err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pages, nil
}
