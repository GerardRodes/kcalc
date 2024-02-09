create table cookings (
	id integer primary key,
	external_id text not null unique,
	name text check(length(name) <= 512),
	user_id integer references users(id) not null,
	g_after_cooking real,
	created_at integer not null,
	updated_at integer not null
);

create table rel_cookings_foods (
	cooking_id integer references cookings(id) not null,
	food_id integer references foods(id) not null,
	g real not null default 0,
	primary key(cooking_id, food_id)
) without rowid;

create table rel_cookings_cookings (
	cooking_id integer references cookings(id) not null,
	sub_cooking_id integer check(sub_cooking_id != cooking_id) references cookings(id) not null,
	g real not null default 0,
	primary key(cooking_id, sub_cooking_id)
) without rowid;