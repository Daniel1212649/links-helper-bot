alter table users
	drop constraint if exists users_locale_check;

alter table users
	drop column if exists locale;
