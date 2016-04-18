package main

import (
	"database/sql"
	"log"
	"time"
)

type User struct {
	ID          int64     `json:"id"`
	Firstname   string    `json:"firstname,omitempty"`
	Lastname    string    `json:"lastname,omitempty"`
	Description string    `json:"description,omitempty"`
	Email       string    `json:"email,omitempty"`
	Password    string    `json:"-"`
	FacebookID  string    `json:"facebook_id,omitempty"`
	InstagramID string    `json:"instagram_id,omitempty"`
	TwitterID   string    `json:"twitter_id,omitempty"`
	ImageURL    string    `json:"image_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func userExists(user *User, fetch bool) (bool, *User) {
	if fetch {
		var firstname, lastname, description, email, password, facebookID, instagramID, twitterID, imageURL sql.NullString

		if err := db.QueryRow(GET_USER_SQL, user.ID, user.Email, user.FacebookID, user.InstagramID, user.TwitterID).Scan(
			&user.ID,
			&firstname,
			&lastname,
			&description,
			&email,
			&password,
			&facebookID,
			&instagramID,
			&twitterID,
			&imageURL,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			if err != sql.ErrNoRows {
				log.Println(err)
			}
			return false, nil
		}

		user.Firstname = fromNullString(firstname)
		user.Lastname = fromNullString(lastname)
		user.Description = fromNullString(description)
		user.Email = fromNullString(email)
		user.Password = fromNullString(password)
		user.FacebookID = fromNullString(facebookID)
		user.InstagramID = fromNullString(instagramID)
		user.TwitterID = fromNullString(twitterID)
		user.ImageURL = fromNullString(imageURL)

		return true, user
	} else {
		count := 0
		if err := db.QueryRow(COUNT_USER_SQL, user.ID, user.Email, user.FacebookID, user.InstagramID, user.TwitterID).Scan(&count); err != nil || count == 0 {
			log.Println(err)
			return false, nil
		}
		return true, nil
	}
}

func insertUser(user *User) error {
	_, err := db.Exec(
		INSERT_USER_SQL,
		&user.Firstname,
		&user.Lastname,
		&user.Description,
		&user.Email,
		&user.Password,
		&user.FacebookID,
		&user.InstagramID,
		&user.TwitterID,
		&user.ImageURL,
	)
	return err
}

func updateUser(user *User) (err error) {
    if user.Email != "" && user.Password != "" {
        _, err = db.Exec(
            UPDATE_USER_SQL,
            &user.Firstname,
            &user.Lastname,
            &user.ImageURL,
        )
    } else {
        _, err = db.Exec(
            UPDATE_USER_WITH_EMAIL_AND_PASSWORD_SQL,
            &user.Firstname,
            &user.Lastname,
            &user.Email,
            &user.Password,
            &user.ImageURL,
        )
    }

	return
}
