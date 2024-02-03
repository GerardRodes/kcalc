create table langs (
	id integer primary key,
	name text not null
);
insert into langs (name) values ('es'), ('es_ES'), ('es_MX'), ('en'), ('en_US'), ('en_GB');

create table users (
	id integer primary key,
	email text check(length(email) <= 255) not null,
	pass_hash blob check(length(pass_hash) <= 255) not null,
	verified_email boolean not null default 'false',
	lang integer references langs(id) not null
);

create table families (
	id integer primary key,
	name text check(length(name) <= 255) not null
);

create table rel_families_users (
	user_id integer references users(id) not null,
	family_id integer references families(id) not null,
	primary key(user_id, family_id)
) without rowid;

create table sources (
	id integer primary key,
	name text not null unique
);

-- name is at locales
create table foods (
	id integer primary key
);

create table foods_locales (
	food_id integer references foods(id) not null,
	lang_id integer references langs(id) not null,
	value text not null check(length(value) <= 512),
	value_normal text not null,
	unique(food_id, lang_id)
); -- needs rowid for fts

create table foods_details (
	food_id integer references foods(id) not null,
	user_id integer references users(id),
	source_id integer references sources(id),
	-- nutrients, all per gram
	kcal real not null,
	unique(food_id, user_id, source_id)
);

create table foods_images (
	food_id integer references foods(id) not null,
	user_id integer references users(id),
	source_id integer references sources(id),
	kind text,
	uri text not null unique
);

