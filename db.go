package main

const CREATE_USER_TABLE_SQL =`
CREATE TABLE "user" (
	id bigserial PRIMARY KEY,
	name varchar(255),
	description varchar(500),
	email varchar(255),
	facebook_id varchar(50),
	instagram_id varchar(50),
	twitter_id varchar(50),
	image_url varchar(255),
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null
)`

const CREATE_HOOP_TABLE_SQL = `
CREATE TABLE hoop (
	id bigserial primary key,
	user_id bigserial not null,
	name varchar(255) not null,
	description varchar(255) not null,
	latitude real not null,
	longitude real not null,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	UNIQUE (name),
	FOREIGN KEY(user_id) REFERENCES "user" (id)
)`

const CREATE_STORY_TABLE_SQL = `
CREATE TABLE story (
	id bigserial primary key,
	user_id bigserial not null,
	name varchar(255) not null,
	description varchar(255) not null,
	image_url varchar(255) not null,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	FOREIGN KEY(user_id) REFERENCES "user" (id)
)`

const CREATE_ACTIVITY_TABLE_SQL = `
CREATE TABLE activity (
	id bigserial primary key,
	user_id bigserial not null,
	type bigint not null,
	hoop_id bigserial,
	story_id bigserial,
	created_at timestamp with time zone not null,
	FOREIGN KEY(user_id) REFERENCES "user" (id)
)`

const CREATE_HOOP_FEATURED_STORY_TABLE_SQL = `
CREATE TABLE hoop_featured_story (
	hoop_id bigserial not null,
	story_id bigserial not null,
	FOREIGN KEY(hoop_id) REFERENCES hoop (id),
	FOREIGN KEY(story_id) REFERENCES story (id)
)`

const CREATE_HOOP_STORY_TABLE_SQL = `
CREATE TABLE hoop_story (
	hoop_id bigserial not null,
	story_id bigserial not null,
	FOREIGN KEY(hoop_id) REFERENCES hoop (id),
	FOREIGN KEY(story_id) REFERENCES story (id)
)`
