package main

import (
    "database/sql"
)

func fromNullString(s sql.NullString) string {
    if s.Valid {
        return s.String
    }
    return ""
}

func fromNullInt64(i sql.NullInt64) int64 {
    if i.Valid {
        return i.Int64
    }
    return 0
}

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
	updated_at timestamp with time zone not null,
    UNIQUE (email, facebook_id, instagram_id, twitter_id)
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
	hoop_id bigserial not null,
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
	hoop_id bigserial primary key,
	story_id bigserial not null,
	FOREIGN KEY(hoop_id) REFERENCES hoop (id),
	FOREIGN KEY(story_id) REFERENCES story (id),
    UNIQUE (story_id)
)`

const CREATE_COMMENT_TABLE_SQL = `
CREATE TABLE comment (
    id bigserial primary key,
    user_id bigserial not null,
    text varchar(255) not null,
	FOREIGN KEY(user_id) REFERENCES "user" (id),
	FOREIGN KEY(hoop_id) REFERENCES hoop (id),
	FOREIGN KEY(story_id) REFERENCES story (id)
)`

// User
const INSERT_USER_SQL = `
INSERT INTO "user" (name, description, email, facebook_id, instagram_id, twitter_id, image_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW()) ON CONFLICT (email, facebook_id, instagram_id, twitter_id) DO NOTHING
RETURNING id`

const GET_USER_WITH_ID_SQL = `
SELECT * FROM "user"
WHERE id = $1
LIMIT 1`

const COUNT_USER_WITH_ID_SQL = `
SELECT COUNT(id) FROM "user"
WHERE id = $1
LIMIT 1`

// Hoop
const INSERT_HOOP_SQL = `
INSERT INTO hoop (user_id, name, description, latitude, longitude, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id`

const GET_HOOPS_SQL = `
SELECT id, user_id, name, description, latitude, longitude, created_at, updated
FROM hoop`

// Story
const INSERT_STORY_SQL = `
INSERT INTO story (hoop_id, user_id, name, description, image_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id`

const GET_STORIES_SQL = `
SELECT id, hoop_id, user_id, name, description, image_url, created_at, updated_at
FROM story
WHERE hoop_id = $1`

// HoopFeaturedStory
const INSERT_HOOP_FEATURED_STORY_SQL = `
INSERT INTO hoop_featured_story (hoop_id, story_id)
VALUES ($1, $2)`

// Activity
const GET_ACTIVITIES_SQL = `
SELECT user_id, type, hoop_id, story_id, created_at FROM activity
LIMIT 100`

const INSERT_POST_HOOP_ACTIVITY_SQL = `
INSERT INTO activity (user_id, type, hoop_id, created_at)
VALUES ($1, $2, $3, NOW())`

const INSERT_POST_STORY_ACTIVITY_SQL = `
INSERT INTO activity (user_id, type, story_id, created_at)
VALUES ($1, $2, $3, NOW())`

const INSERT_HOOP_COMMENT_ACTIVITY_SQL = `
INSERT INTO activity (user_id, type, hoop_id, created_at)
VALUES ($1, $2, $3, NOW())`

const INSERT_STORY_COMMENT_ACTIVITY_SQL = `
INSERT INTO activity (user_id, type, story_id, created_at)
VALUES ($1, $2, $3, NOW())`

const INSERT_HOOP_LIKE_ACTIVITY_SQL = `
INSERT INTO activity (user_id, type, hoop_id, created_at)
VALUES ($1, $2, $3, NOW())`

const INSERT_STORY_LIKE_ACTIVITY_SQL = `
INSERT INTO activity (user_id, type, story_id, created_at)
VALUES ($1, $2, $3, NOW())`

// Comment
const GET_HOOP_COMMENTS_SQL = `
SELECT user_id, hoop_id, text, created_at, updated_at FROM comment`

const GET_STORY_COMMENTS_SQL = `
SELECT user_id, story_id, text, created_at, updated_at FROM comment`

const INSERT_HOOP_COMMENT_SQL = `
INSERT INTO comment (user_id, text, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
RETURNING id`

const INSERT_STORY_COMMENT_SQL = `
INSERT INTO comment (user_id, text, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
RETURNING id`

// Like
const GET_HOOP_LIKES_SQL = `
SELECT user_id, hoop_id, created_at, updated_at FROM comment`

const GET_STORY_LIKES_SQL = `
SELECT user_id, story_id, created_at, updated_at FROM comment`

const INSERT_HOOP_LIKE_SQL = `
INSERT INTO comment (user_id, created_at, updated_at)
VALUES ($1, NOW(), NOW())
RETURNING id`

const INSERT_STORY_LIKE_SQL = `
INSERT INTO comment (user_id, created_at, updated_at)
VALUES ($1, NOW(), NOW())
RETURNING id`
