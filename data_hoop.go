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
