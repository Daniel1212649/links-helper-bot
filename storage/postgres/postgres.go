package postgres

import (
	"context"
	"errors"

	"github.com/Daniel1212649/LinksHelperBot/storage"
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
	normalizedURL, err := storage.NormalizeURL(rawURL)
	if err != nil {
		return nil, err
	}

	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return nil, err
	}

	const query = `
		insert into links (user_id, url, normalized_url, status)
		values ($1, $2, $3, $4)
		returning id, url, normalized_url, coalesce(title, ''), coalesce(description, ''), status, created_at, updated_at, read_at`

	var page storage.Page
	err = s.pool.QueryRow(ctx, query, userID, normalizedURL, normalizedURL, storage.StatusUnread).Scan(
		&page.ID,
		&page.URL,
		&page.NormalizedURL,
		&page.Title,
		&page.Description,
		&page.Status,
		&page.CreatedAt,
		&page.UpdatedAt,
		&page.ReadAt,
	)
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
		select id, url, normalized_url, coalesce(title, ''), coalesce(description, ''), status, created_at, updated_at, read_at
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
		&page.Status,
		&page.CreatedAt,
		&page.UpdatedAt,
		&page.ReadAt,
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
		select id, url, normalized_url, coalesce(title, ''), coalesce(description, ''), status, created_at, updated_at, read_at
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

func (s *Storage) Search(ctx context.Context, user storage.User, queryText string, limit int) ([]storage.Page, error) {
	userID, err := s.ensureUser(ctx, user)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	const query = `
		select id, url, normalized_url, coalesce(title, ''), coalesce(description, ''), status, created_at, updated_at, read_at
		from links
		where user_id = $1
			and status <> $2
			and (url ilike '%' || $3 || '%' or coalesce(title, '') ilike '%' || $3 || '%')
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

func (s *Storage) ensureUser(ctx context.Context, user storage.User) (int64, error) {
	const query = `
		insert into users (telegram_user_id, username, chat_id)
		values (nullif($1, 0), nullif($2, ''), $3)
		on conflict (chat_id) do update
		set telegram_user_id = coalesce(excluded.telegram_user_id, users.telegram_user_id),
			username = coalesce(excluded.username, users.username),
			updated_at = now()
		returning id`

	var id int64
	err := s.pool.QueryRow(ctx, query, user.TelegramID, user.Username, user.ChatID).Scan(&id)
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
			&page.Status,
			&page.CreatedAt,
			&page.UpdatedAt,
			&page.ReadAt,
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
