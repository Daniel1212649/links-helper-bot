drop index if exists links_due_reminders_idx;

alter table links
	drop column if exists reminded_at,
	drop column if exists remind_at,
	drop column if exists note;
