alter table users
	add column if not exists locale text not null default 'ru';

alter table users
	add constraint users_locale_check check (locale in ('ru', 'en'));
