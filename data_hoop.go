package main

import (
	"database/sql"
	"log"
	"time"
)

type Hoop struct {
	ID          int64                  `json:"id"`
	UserID      int64                  `json:"user_id"`
	User        User                   `json:"user"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Latitude    float64                `json:"latitude"`
	Longitude   float64                `json:"longitude"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

func hoopExists(hoop *Hoop, fetch bool) (bool, *Hoop) {
	if fetch {
		if newHoop, err := getHoop(hoop.ID); err != nil {
			log.Println(err)
			return false, nil
		} else {
			*hoop = newHoop
		}

		return true, hoop
	} else {
		count := 0

		if err := db.QueryRow(COUNT_HOOP_SQL, hoop.ID).Scan(&count); err != nil || count == 0 {
			log.Println(err)
			return false, nil
		}

		return true, nil
	}
}

func getHoop(hoopID int64) (hoop Hoop, err error) {
	if err = db.QueryRow(GET_HOOP_SQL, hoopID).Scan(
		&hoop.ID,
		&hoop.UserID,
		&hoop.Name,
		&hoop.Description,
		&hoop.Latitude,
		&hoop.Longitude,
		&hoop.CreatedAt,
		&hoop.UpdatedAt,
	); err != nil {
		return
	}

	if hoop.User, err = getUserByID(hoop.UserID); err != nil {
		return
	}

	var featuredStory Story
	if featuredStory, err = getFeaturedStory(hoop.ID); err != nil {
		return
	} else {
		hoop.Data = map[string]interface{}{}
		hoop.Data["featured_story"] = featuredStory
	}

	return
}

func getHoops(query string, args ...interface{}) (hoops []Hoop, err error) {
	var rows *sql.Rows

	if rows, err = db.Query(query, args...); err != nil {
		return
	}

	for rows.Next() {
		var hoop Hoop

		if err = rows.Scan(
			&hoop.ID,
			&hoop.UserID,
			&hoop.Name,
			&hoop.Description,
			&hoop.Latitude,
			&hoop.Longitude,
			&hoop.CreatedAt,
			&hoop.UpdatedAt,
		); err != nil {
			return
		}

		if hoop.User, err = getUserByID(hoop.UserID); err != nil {
			return
		}

		var featuredStory Story
		if featuredStory, err = getFeaturedStory(hoop.ID); err != nil {
			return
		} else {
			hoop.Data = map[string]interface{}{}
			hoop.Data["featured_story"] = featuredStory
		}

		hoops = append(hoops, hoop)
	}

	return
}

func insertHoop(userID int64, name, description, hoopImageURL, courtImageURL, crewImageURL string, latitude, longitude float64) error {
	// Start Transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	var hoopID, storyID int64

	// Insert Hoop
	if err := tx.QueryRow(INSERT_HOOP_SQL, userID, name, description, latitude, longitude).Scan(&hoopID); err != nil {
		return err
	}

	// Insert Hoop Story
	if hoopImageURL != "" {
		if err := tx.QueryRow(INSERT_STORY_SQL, hoopID, userID, name, description, hoopImageURL).Scan(&storyID); err != nil {
			return err
		}

		if _, err := tx.Exec(INSERT_HOOP_FEATURED_STORY_SQL, hoopID, storyID, "hoop"); err != nil {
			return err
		}
	}

	// Insert Court Story
	if courtImageURL != "" {
		if err := tx.QueryRow(INSERT_STORY_SQL, hoopID, userID, name, description, courtImageURL).Scan(&storyID); err != nil {
			return err
		}

		if _, err := tx.Exec(INSERT_HOOP_FEATURED_STORY_SQL, hoopID, storyID, "court"); err != nil {
			return err
		}
	}

	// Insert Crew Story
	if crewImageURL != "" {
		if err := tx.QueryRow(INSERT_STORY_SQL, hoopID, userID, name, description, crewImageURL).Scan(&storyID); err != nil {
			return err
		}

		if _, err := tx.Exec(INSERT_HOOP_FEATURED_STORY_SQL, hoopID, storyID, "crew"); err != nil {
			return err
		}
	}

	// Insert Activity
	if _, err := tx.Exec(INSERT_POST_HOOP_ACTIVITY_SQL, userID, ACTIVITY_POST_HOOP, hoopID); err != nil {
		return err
	}

	// End Transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
