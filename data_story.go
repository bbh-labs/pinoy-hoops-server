package main

import (
	"database/sql"
	"log"
	"time"
)

type Story struct {
	ID        int64     `json:"id"`
	HoopID    int64     `json:"hoop_id"`
	UserID    int64     `json:"user_id"`
	Hoop      Hoop      `json:"hoop"`
	User      User      `json:"user"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func storyExists(story *Story, fetch bool) (bool, *Story) {
	if fetch {
		var imageURL sql.NullString

		if err := db.QueryRow(GET_STORY_SQL, story.ID).Scan(
			&story.ID,
			&story.HoopID,
			&story.UserID,
			&imageURL,
			&story.CreatedAt,
			&story.UpdatedAt,
		); err != nil {
			log.Println(err)
			return false, nil
		}

		story.ImageURL = fromNullString(imageURL)

		return true, story
	} else {
		count := 0
		if err := db.QueryRow(COUNT_STORY_SQL, story.ID).Scan(&count); err != nil || count == 0 {
			log.Println(err)
			return false, nil
		}
		return true, nil
	}
}

func getStory(storyID int64) (story Story, err error) {
	if err = db.QueryRow(GET_STORY_SQL, storyID).Scan(
		&story.ID,
		&story.HoopID,
		&story.UserID,
		&story.ImageURL,
		&story.CreatedAt,
		&story.UpdatedAt,
	); err != nil {
		return
	}

	if story.Hoop, err = getHoop(story.HoopID); err != nil {
		return
	}

	if story.User, err = getUserByID(story.UserID); err != nil {
		return
	}

	return
}

func getFeaturedStories(hoopID int64) (stories map[string]Story, err error) {
	rows, err := db.Query(GET_FEATURED_STORIES_SQL, hoopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var story Story
		var typ string

		if err = rows.Scan(
			&story.ID,
			&story.HoopID,
			&story.UserID,
			&story.ImageURL,
			&story.CreatedAt,
			&story.UpdatedAt,
			&typ,
		); err != nil {
			return
		}

		if story.User, err = getUserByID(story.UserID); err != nil {
			return
		}

		if stories == nil {
			stories = make(map[string]Story)
		}

		stories[typ] = story
	}

	return
}

func getStories(query string, hoopID int64) ([]Story, error) {
	var stories []Story

	rows, err := db.Query(query, hoopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var story Story

		if err := rows.Scan(
			&story.ID,
			&story.HoopID,
			&story.UserID,
			&story.ImageURL,
			&story.CreatedAt,
			&story.UpdatedAt,
		); err != nil {
			return nil, err
		}

		stories = append(stories, story)
	}

	return stories, nil
}

func insertStory(hoopID, userID int64, imageURL string) error {
	var storyID int64

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Insert Story
	if err := tx.QueryRow(INSERT_STORY_SQL, hoopID, userID, imageURL).Scan(&storyID); err != nil {
		return err
	}

	// Insert Activity
	if _, err := tx.Exec(INSERT_POST_STORY_ACTIVITY_SQL, userID, ACTIVITY_POST_STORY, storyID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
