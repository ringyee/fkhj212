-- +migrate Up
create table if not exists app_user (
	id bigserial primary key,
	username character varying (100) not null,
	password_hash character varying (200) not null,
	session_ttl bigint not null,
	is_admin boolean not null
);

create unique index idx_user_username on app_user(username);

-- global admin (password: admin)
insert into app_user (
	username,
	password_hash,
	session_ttl,
	is_admin
) values (
	'admin',
	'PBKDF2$sha512$100000$4u3hL8krvlMIS0KnCYXeMw==$G7c7tuUYq2zSJaUeruvNL/KF30d3TVDORVD56wzvJYmc3muWjoaozH8bHJ7r8zY8dW6Pts2bWyhFfkb/ubQZsA==',
	0,
	true
);


-- +migrate Down
drop index idx_user_username;
drop table app_user;
