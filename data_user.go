package main

import (
	"database/sql"
    "fmt"
    "log"
	"time"

    "github.com/garyburd/redigo/redis"
)

type User struct {
	ID          int64                 `json:"id"`
	Firstname   string                `json:"firstname,omitempty"`
	Lastname    string                `json:"lastname,omitempty"`
	Description string                `json:"description,omitempty"`
	Email       string                `json:"email,omitempty"`
	Password    string                `json:"-"`
	FacebookID  string                `json:"facebook_id,omitempty"`
	InstagramID string                `json:"instagram_id,omitempty"`
	TwitterID   string                `json:"twitter_id,omitempty"`
	ImageURL    string                `json:"image_url,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
    LatestActivityCheckTime time.Time `json:"latest_activity_check_time,omitempty"`
}

func (user *User) lastActivityCheckTime() (time.Time, error) {
    if reply, err := red.Do("HGET", fmt.Sprintf("user:%d", user.ID), "lastActivityCheckTime"); err != nil {
        return time.Time{}, err
    } else if t, err := redis.Int64(reply, err); err != nil {
        if err != redis.ErrNil {
            return time.Time{}, err
        }
    } else {
        return time.Unix(t, 0), nil
    }
    return time.Time{}, nil
}

func (user *User) updateLastActivityCheckTime(secs int64) error {
    if _, err := red.Do("HSET", fmt.Sprintf("user:%d", user.ID), "lastActivityCheckTime", secs); err != nil {
        return err
    }
    return nil
}

func userExists(user *User, fetch bool) (bool, *User) {
    var err error

	if fetch {
		var firstname, lastname, description, email, password, facebookID, instagramID, twitterID, imageURL sql.NullString

		if err = db.QueryRow(GET_USER_SQL, user.ID, user.Email, user.FacebookID, user.InstagramID, user.TwitterID).Scan(
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
        if user.LatestActivityCheckTime, err = user.lastActivityCheckTime(); err != nil {
            log.Println(err)
            return false, nil
        }

		return true, user
	} else {
		count := 0
		if err = db.QueryRow(COUNT_USER_SQL, user.ID, user.Email, user.FacebookID, user.InstagramID, user.TwitterID).Scan(&count); err != nil || count == 0 {
			log.Println(err)
			return false, nil
		}
		return true, nil
	}
}

func getUserByID(userID int64) (User, error) {
    var user User
    var firstname, lastname, description, email, password, facebookID, instagramID, twitterID, imageURL sql.NullString
    var err error

    if err = db.QueryRow(GET_USER_BY_ID_SQL, userID).Scan(
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
        return user, err
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
    if user.LatestActivityCheckTime, err = user.lastActivityCheckTime(); err != nil {
        log.Println(err)
        return user, nil
    }

    return user, nil
}

func insertUser(user *User) (int64, error) {
    var userID int64

	if err := db.QueryRow(
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
	).Scan(&userID); err != nil && err != sql.ErrNoRows {
        return 0, err
    }

	return userID, nil
}

func updateUser(user *User) (err error) {
    if user.Email != "" && user.Password != "" {
        _, err = db.Exec(
            UPDATE_USER_WITH_EMAIL_AND_PASSWORD_SQL,
            &user.Firstname,
            &user.Lastname,
            &user.Email,
            &user.Password,
            &user.ImageURL,
        )
    } else {
        _, err = db.Exec(
            UPDATE_USER_SQL,
            &user.Firstname,
            &user.Lastname,
            &user.ImageURL,
        )
    }

	return
}

func view(otherID int64, typ string) error {
    if _, err := red.Do("HINCRBY", fmt.Sprintf("%s:%d", typ, otherID), "view_count", 1); err != nil {
        return err
    }
    return nil
}
