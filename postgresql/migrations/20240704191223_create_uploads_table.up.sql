begin;

create type upload_status as enum ('in_progress', 'done');

create table uploads (
	id uuid primary key,
	name text not null,
	size bigint not null,
	status upload_status not null,
	created_at timestamptz not null default now()
);

create table parts (
	id uuid primary key,
	server_url text not null,
	upload_id uuid not null references uploads(id) ON DELETE CASCADE,
	number int not null,
	size bigint not null,
	created_at timestamptz not null default now(),
	status upload_status not null
);

commit;
