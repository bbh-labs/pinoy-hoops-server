package main

import (
	"database/sql"
	"log"
	"time"
)

type Hoop struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func hoopExists(hoop *Hoop, fetch bool) (bool, *Hoop) {
	if fetch {
		var name, description sql.NullString

		if err := db.QueryRow(GET_HOOP_SQL, hoop.ID).Scan(
			&hoop.ID,
			&hoop.UserID,
			&name,
			&description,
			&hoop.Latitude,
			&hoop.Longitude,
			&hoop.CreatedAt,
			&hoop.UpdatedAt,
		); err != nil {
			log.Println(err)
			return false, nil
		}

		hoop.Name = fromNullString(name)
		hoop.Description = fromNullString(description)

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

func getHoops(name string) ([]Hoop, error) {
    var hoops []Hoop
    var rows *sql.Rows
    var err error

    if name != "" {
        rows, err = db.Query(GET_HOOPS_WITH_NAME_SQL, "%"+name+"%")
    } else {
        rows, err = db.Query(GET_HOOPS_SQL)
    }

    if err != nil {
        return nil, err
    }

    for rows.Next() {
        var hoop Hoop

        if err := rows.Scan(
            &hoop.ID,
            &hoop.UserID,
            &hoop.Name,
            &hoop.Description,
            &hoop.Latitude,
            &hoop.Longitude,
            &hoop.CreatedAt,
            &hoop.UpdatedAt,
        ); err != nil {
            return nil, err
        }

        hoops = append(hoops, hoop)
    }

    return hoops, nil
}

func insertHoop(userID int64, name, description, imageURL string, latitude, longitude float64) error {
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

    // Insert Story
    if err := tx.QueryRow(INSERT_STORY_SQL, hoopID, userID, name, description, imageURL).Scan(&storyID); err != nil {
        return err
    }

    // Insert HoopFeaturedStory
    if _, err := tx.Exec(INSERT_HOOP_FEATURED_STORY_SQL, hoopID, storyID); err != nil {
        return err
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
