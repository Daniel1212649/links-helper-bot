alter table links
	add column if not exists note text not null default '',
	add column if not exists remind_at timestamptz,
	add column if not exists reminded_at timestamptz;

create index if not exists links_due_reminders_idx
	on links (remind_at)
	where remind_at is not null and reminded_at is null and status <> 'deleted';
