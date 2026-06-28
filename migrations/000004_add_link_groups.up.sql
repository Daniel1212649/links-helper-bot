alter table links
	add column if not exists group_name text not null default '';

create index if not exists links_user_group_created_at_idx
	on links (user_id, lower(group_name), created_at desc)
	where status <> 'deleted' and group_name <> '';
