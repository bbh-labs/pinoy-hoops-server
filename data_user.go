package main

import (
    "database/sql"
    "log"
    "time"
)

type User struct {
    ID int64 `json:"id"`
    Name sql.NullString `json:"name"`
    Description sql.NullString `json:"description"`
    Email sql.NullString `json:"email"`
    FacebookID sql.NullString `json:"facebook_id"`
    InstagramID sql.NullString `json:"instagram_id"`
    TwitterID sql.NullString `json:"twitter_id"`
    ImageURL sql.NullString `json:"image_url"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func userExists(user *User, fetchUser bool) (bool, *User) {
    if fetchUser {
        user := &User{}
        if err := db.QueryRow(`SELECT * FROM "user" WHERE id = $1 LIMIT 1`, user.ID).Scan(
            &user.ID,
            &user.Name,
            &user.Email,
            &user.FacebookID,
            &user.InstagramID,
            &user.TwitterID,
            &user.ImageURL,
            &user.CreatedAt,
            &user.UpdatedAt,
        ); err != nil {
            log.Println(err)
            return false, nil
        }
        return true, user
    } else {
        count := 0
        if err := db.QueryRow(`SELECT COUNT(id) FROM "user" WHERE id = $1 LIMIT 1`, user.ID).Scan(&count); err != nil {
            log.Println(err)
            return false, nil
        }
        return true, nil
    }
}

func insertUser(user *User) error {
    _, err := db.Exec(
        `INSERT INTO "user" VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`,
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
