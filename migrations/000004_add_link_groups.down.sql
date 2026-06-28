drop index if exists links_user_group_created_at_idx;

alter table links
	drop column if exists group_name;
