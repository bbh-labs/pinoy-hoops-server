package main

import (
	"database/sql"
	"log"
	"time"
)

type User struct {
	ID                      int64     `json:"id"`
	Firstname               string    `json:"firstname,omitempty"`
	Lastname                string    `json:"lastname,omitempty"`
	Gender                  string    `json:"gender,omitempty"`
	Birthdate               string    `json:"birthdate,omitempty"`
	Description             string    `json:"description,omitempty"`
	Email                   string    `json:"email,omitempty"`
	Password                string    `json:"-"`
	FacebookID              string    `json:"facebook_id,omitempty"`
	ImageURL                string    `json:"image_url,omitempty"`
	BackgroundURL           string    `json:"background_url,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

func (user *User) updateUserImage(imageURL string) (err error) {
	_, err = db.Exec(UPDATE_USER_IMAGE_SQL, imageURL, user.ID)
	return
}

func (user *User) updateBackgroundImage(backgroundURL string) (err error) {
	_, err = db.Exec(UPDATE_USER_BACKGROUND_SQL, backgroundURL, user.ID)
	return
}

func userExists(user *User, fetch bool) (bool, *User) {
	var err error

	if fetch {
		var firstname, lastname, gender, birthdate, description, email, password, facebookID, imageURL, backgroundURL sql.NullString

		if err = db.QueryRow(GET_USER_SQL, user.ID, user.Email, user.FacebookID).Scan(
			&user.ID,
			&firstname,
			&lastname,
			&gender,
			&birthdate,
			&description,
			&email,
			&password,
			&facebookID,
			&imageURL,
			&backgroundURL,
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
		user.Gender = fromNullString(gender)
		user.Birthdate = fromNullString(birthdate)
		user.Description = fromNullString(description)
		user.Email = fromNullString(email)
		user.Password = fromNullString(password)
		user.FacebookID = fromNullString(facebookID)
		user.ImageURL = fromNullString(imageURL)
		user.BackgroundURL = fromNullString(backgroundURL)

		return true, user
	} else {
		count := 0
		if err = db.QueryRow(COUNT_USER_SQL, user.ID, user.Email, user.FacebookID).Scan(&count); err != nil || count == 0 {
			log.Println(err)
			return false, nil
		}
		return true, nil
	}
}

func getUserByID(userID int64) (User, error) {
	var user User
	var firstname, lastname, gender, birthdate, description, email, password, facebookID, imageURL, backgroundURL sql.NullString
	var err error

	if err = db.QueryRow(GET_USER_BY_ID_SQL, userID).Scan(
		&user.ID,
		&firstname,
		&lastname,
		&gender,
		&birthdate,
		&description,
		&email,
		&password,
		&facebookID,
		&imageURL,
		&backgroundURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return user, err
	}

	user.Firstname = fromNullString(firstname)
	user.Lastname = fromNullString(lastname)
	user.Gender = fromNullString(gender)
	user.Birthdate = fromNullString(birthdate)
	user.Description = fromNullString(description)
	user.Email = fromNullString(email)
	user.Password = fromNullString(password)
	user.FacebookID = fromNullString(facebookID)
	user.ImageURL = fromNullString(imageURL)
	user.BackgroundURL = fromNullString(backgroundURL)

	return user, nil
}

func insertUser(user *User) (int64, error) {
	var userID int64

	if err := db.QueryRow(
		INSERT_USER_SQL,
		&user.Firstname,
		&user.Lastname,
		&user.Gender,
		&user.Birthdate,
		&user.Description,
		&user.Email,
		&user.Password,
		&user.FacebookID,
		&user.ImageURL,
	).Scan(&userID); err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	return userID, nil
}

func updateUser(user *User) (err error) {
	_, err = db.Exec(
		UPDATE_USER_SQL,
		&user.Firstname,
		&user.Lastname,
		&user.Gender,
		&user.Birthdate,
		&user.ID,
	)

	return
}
