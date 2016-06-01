package main

import (
	"time"
)

type Like struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toggleLike(userID int64, otherID int64, typ string) error {
	var query string
	var activity int
	var err error

	switch typ {
	case "hoop":
		query = COUNT_HOOP_ACTIVITY_BY_USER_SQL
		activity = ACTIVITY_POST_LIKE_HOOP
	case "story":
		query = COUNT_STORY_ACTIVITY_BY_USER_SQL
		activity = ACTIVITY_POST_LIKE_STORY
	}

	// Check if user liked before
	var count int64
	if err := db.QueryRow(query, userID, activity, otherID).Scan(&count); err != nil {
		return err
	} else {
		if count > 0 {
			if err := deleteLike(userID, otherID, typ); err != nil {
				return err
			}
			return nil
		}
	}

	switch typ {
	case "hoop":
		query = INSERT_HOOP_LIKE_ACTIVITY_SQL
	case "story":
		query = INSERT_STORY_LIKE_ACTIVITY_SQL
	}

	// Insert Activity
	if _, err = db.Exec(query, userID, activity, otherID); err != nil {
		return err
	}

	return nil
}

func deleteLike(userID int64, otherID int64, typ string) error {
	var query string
	var activity int
	var err error

	switch typ {
	case "hoop":
		query = DELETE_HOOP_ACTIVITY_SQL
		activity = ACTIVITY_POST_LIKE_HOOP
	case "story":
		query = DELETE_STORY_ACTIVITY_SQL
		activity = ACTIVITY_POST_LIKE_STORY
	}

	// Delete Activity
	if _, err = db.Exec(query, userID, activity, otherID); err != nil {
		return err
	}

	return nil
}
