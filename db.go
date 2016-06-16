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

const CREATE_USER_TABLE_SQL = `
CREATE TABLE "user" (
	id bigserial PRIMARY KEY,
	firstname varchar(255),
	lastname varchar(255),
	gender varchar(6),
	birthdate varchar(16),
	description varchar(500),
	email varchar(255),
	password varchar(60),
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
	FOREIGN KEY(user_id) REFERENCES "user" (id),
	FOREIGN KEY(hoop_id) REFERENCES hoop (id)
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
	type text not null,
	FOREIGN KEY(hoop_id) REFERENCES hoop (id),
	FOREIGN KEY(story_id) REFERENCES story (id),
    UNIQUE (story_id)
)`

const CREATE_COMMENT_TABLE_SQL = `
CREATE TABLE comment (
    id bigserial primary key,
    user_id bigserial not null,
    text varchar(255) not null,
    hoop_id bigserial,
    story_id bigserial,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
	FOREIGN KEY(user_id) REFERENCES "user" (id)
)`

// User
const INSERT_USER_SQL = `
INSERT INTO "user" (firstname, lastname, gender, birthdate, description, email, password, facebook_id, instagram_id, twitter_id, image_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
ON CONFLICT (email, facebook_id, instagram_id, twitter_id)
DO NOTHING
RETURNING id`

const UPDATE_USER_SQL = `
UPDATE "user" SET
firstname = $1,
lastname = $2,
gender = $3,
birthdate = $4,
updated_at = NOW()
WHERE id = $5`

const UPDATE_USER_WITH_EMAIL_AND_PASSWORD_SQL = `
UPDATE "user" SET
firstname = $1,
lastname = $2,
email = $3,
password = $4,
image_url = $5,
updated_at = NOW()`

const UPDATE_USER_FACEBOOK_SQL = `
UPDATE "user" SET facebook_id = $1 WHERE id = $2`

const UPDATE_USER_INSTAGRAM_SQL = `
UPDATE "user" SET instagram_id = $1 WHERE id = $2`

const UPDATE_USER_TWITTER_SQL = `
UPDATE "user" SET twitter_id = $1 WHERE id = $2`

const UPDATE_USER_IMAGE_SQL = `
UPDATE "user" SET image_url = $1 WHERE id = $2`

const GET_USER_SQL = `
SELECT id, firstname, lastname, gender, birthdate, description, email, password, facebook_id, instagram_id, twitter_id, image_url, created_at, updated_at FROM "user"
WHERE id = $1
OR (email = $2 AND email != '')
OR (facebook_id = $3 AND facebook_id != '')
OR (instagram_id = $4 AND instagram_id != '')
OR (twitter_id = $5 AND twitter_id != '')
LIMIT 1`

const GET_USER_BY_ID_SQL = `
SELECT id, firstname, lastname, gender, birthdate, description, email, password, facebook_id, instagram_id, twitter_id, image_url, created_at, updated_at FROM "user"
WHERE id = $1
LIMIT 1`

const COUNT_USER_SQL = `
SELECT COUNT(id) FROM "user"
WHERE id = $1 OR email = $2 OR facebook_id = $3 OR instagram_id = $4 OR twitter_id = $5
LIMIT 1`

// Hoop
const INSERT_HOOP_SQL = `
INSERT INTO hoop (user_id, name, description, latitude, longitude, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id`

const GET_HOOP_SQL = `
SELECT id, user_id, name, description, latitude, longitude, created_at, updated_at FROM hoop
WHERE id = $1
LIMIT 1`

const COUNT_HOOP_SQL = `
SELECT COUNT(id) FROM hoop
WHERE id = $1
LIMIT 1`

const GET_HOOPS_SQL = `
SELECT id, user_id, name, description, latitude, longitude, created_at, updated_at
FROM hoop`

const GET_MY_HOOPS_SQL = `
SELECT id, user_id, name, description, latitude, longitude, created_at, updated_at
FROM hoop
WHERE user_id = $1`

const GET_OTHER_HOOPS_SQL = `
SELECT id, user_id, name, description, latitude, longitude, created_at, updated_at
FROM hoop
WHERE user_id != $1`

const DISTANCE_CALC = `(acos(sin(radians(h.latitude)) * sin(radians($1)) + cos(radians(h.latitude)) * cos(radians($2)) * cos(radians(h.longitude - $3))) * 6371 * 1000)`

const GET_NEARBY_HOOPS_SQL = `
SELECT * FROM (SELECT ` + DISTANCE_CALC + ` computedDistance, * FROM hoop h) AS tempQuery WHERE computedDistance < $4 ORDER BY computedDistance ASC LIMIT 100`

const GET_POPULAR_HOOPS_SQL = `
SELECT id, user_id, name, description, latitude, longitude, created_at, updated_at
FROM hoop
ORDER BY (SELECT COUNT(id) FROM story WHERE hoop_id = hoop.id) DESC
LIMIT 100`

const GET_LATEST_HOOPS_SQL = `
SELECT id, user_id, name, description, latitude, longitude, created_at, updated_at
FROM hoop
ORDER BY created_at DESC
LIMIT 100`

const GET_HOOPS_WITH_NAME_SQL = `
SELECT id, user_id, name, description, latitude, longitude, created_at, updated_at
FROM hoop
WHERE name LIKE $1`

// Story
const INSERT_STORY_SQL = `
INSERT INTO story (hoop_id, user_id, name, description, image_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id`

const GET_STORY_SQL = `
SELECT id, hoop_id, user_id, name, description, image_url, created_at, updated_at FROM story
WHERE id = $1
LIMIT 1`

const GET_FEATURED_STORY_SQL = `
SELECT id, hoop_id, user_id, name, description, image_url, created_at, updated_at FROM story
WHERE hoop_id = $1
LIMIT 1`

const COUNT_STORY_SQL = `
SELECT COUNT(id) FROM story
WHERE id = $1
LIMIT 1`

const GET_STORIES_SQL = `
SELECT id, hoop_id, user_id, name, description, image_url, created_at, updated_at
FROM story
WHERE hoop_id = $1`

const GET_MOST_COMMENTED_STORIES_SQL = `
SELECT id, hoop_id, user_id, name, description, image_url, created_at, updated_at
FROM story
WHERE hoop_id = $1
ORDER BY (SELECT COUNT(id) FROM activity WHERE type = 102 AND story_id = story.id) DESC`

const GET_MOST_LIKED_STORIES_SQL = `
SELECT id, hoop_id, user_id, name, description, image_url, created_at, updated_at
FROM story
WHERE hoop_id = $1
ORDER BY (SELECT COUNT(id) FROM activity WHERE type = 202 AND story_id = story.id) DESC`

const GET_LATEST_STORIES_SQL = `
SELECT id, hoop_id, user_id, name, description, image_url, created_at, updated_at
FROM story
WHERE hoop_id = $1
ORDER BY created_at DESC`

// HoopFeaturedStory
const INSERT_HOOP_FEATURED_STORY_SQL = `
INSERT INTO hoop_featured_story (hoop_id, story_id, type)
VALUES ($1, $2, $3)`

// Activity
const GET_ACTIVITIES_SQL = `
SELECT user_id, type, hoop_id, story_id, created_at FROM activity
WHERE user_id != $1
ORDER BY created_at DESC
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
INSERT INTO activity (user_id, type, hoop_id, story_id, created_at)
VALUES ($1, $2, $3, 0, NOW())`

const INSERT_STORY_LIKE_ACTIVITY_SQL = `
INSERT INTO activity (user_id, type, hoop_id, story_id, created_at)
VALUES ($1, $2, 0, $3, NOW())`

const DELETE_HOOP_ACTIVITY_SQL = `
DELETE FROM activity WHERE user_id = $1 AND type = $2 AND hoop_id = $3`

const DELETE_STORY_ACTIVITY_SQL = `
DELETE FROM activity WHERE user_id = $1 AND type = $2 AND story_id = $3`

// Comment
const GET_HOOP_COMMENTS_SQL = `
SELECT id, user_id, text, hoop_id, created_at, updated_at FROM comment
WHERE hoop_id = $1
ORDER BY created_at DESC`

const GET_STORY_COMMENTS_SQL = `
SELECT id, user_id, text, story_id, created_at, updated_at FROM comment
WHERE story_id = $1
ORDER BY created_at DESC`

const INSERT_HOOP_COMMENT_SQL = `
INSERT INTO comment (user_id, text, hoop_id, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id`

const INSERT_STORY_COMMENT_SQL = `
INSERT INTO comment (user_id, text, story_id, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id`

// Like
const COUNT_HOOP_LIKES_SQL = `
SELECT COUNT(id) FROM activity WHERE hoop_id = $1 AND type = 202`

const COUNT_STORY_LIKES_SQL = `
SELECT COUNT(id) FROM activity WHERE story_id = $1 AND type = 202`

const COUNT_HOOP_ACTIVITY_BY_USER_SQL = `
SELECT COUNT(id) FROM activity WHERE user_id = $1 AND type = $2 AND hoop_id = $3`

const COUNT_STORY_ACTIVITY_BY_USER_SQL = `
SELECT COUNT(id) FROM activity WHERE user_id = $1 AND type = $2 AND story_id = $3`
