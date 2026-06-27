create table if not exists users (
	id bigserial primary key,
	telegram_user_id bigint,
	username text,
	chat_id bigint not null unique,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create table if not exists links (
	id bigserial primary key,
	user_id bigint not null references users(id) on delete cascade,
	url text not null,
	normalized_url text not null,
	title text,
	description text,
	status text not null default 'unread',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	read_at timestamptz,
	constraint links_status_check check (status in ('unread', 'read', 'deleted')),
	constraint links_user_normalized_url_unique unique (user_id, normalized_url)
);

create index if not exists links_user_status_created_at_idx
	on links (user_id, status, created_at desc);

create index if not exists links_user_created_at_idx
	on links (user_id, created_at desc);
