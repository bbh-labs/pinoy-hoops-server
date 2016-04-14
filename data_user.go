package main

import (
    "database/sql"
    "log"
    "time"
)

type User struct {
    ID int64 `json:"id"`
    Name string `json:"name,omitempty"`
    Description string `json:"description,omitempty"`
    Email string `json:"email,omitempty"`
    FacebookID string `json:"facebook_id,omitempty"`
    InstagramID string `json:"instagram_id,omitempty"`
    TwitterID string `json:"twitter_id,omitempty"`
    ImageURL string `json:"image_url,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func userExists(user *User, fetchUser bool) (bool, *User) {
    if fetchUser {
        var name, description, email, facebookID, instagramID, twitterID, imageURL sql.NullString

        user := &User{}
        if err := db.QueryRow(GET_USER_WITH_ID_SQL, user.ID).Scan(
            &user.ID,
            &name,
            &description,
            &email,
            &facebookID,
            &instagramID,
            &twitterID,
            &imageURL,
            &user.CreatedAt,
            &user.UpdatedAt,
        ); err != nil {
            log.Println(err)
            return false, nil
        }

        user.Name = fromNullString(name)
        user.Description = fromNullString(description)
        user.Email = fromNullString(email)
        user.FacebookID = fromNullString(facebookID)
        user.InstagramID = fromNullString(instagramID)
        user.TwitterID = fromNullString(twitterID)
        user.ImageURL = fromNullString(imageURL)

        return true, user
    } else {
        count := 0
        if err := db.QueryRow(COUNT_USER_WITH_ID_SQL, user.ID).Scan(&count); err != nil {
            log.Println(err)
            return false, nil
        }
        return true, nil
    }
}

func insertUser(user *User) error {
    _, err := db.Exec(
        INSERT_USER_SQL,
        &user.Name,
        &user.Description,
        &user.Email,
        &user.FacebookID,
        &user.InstagramID,
        &user.TwitterID,
        &user.ImageURL,
    )
    return err
}
