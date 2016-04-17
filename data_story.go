package main

import (
    "database/sql"
    "log"
    "time"
)

type Story struct {
    ID int64 `json:"id"`
    HoopID int64 `json:"hoop_id"`
    UserID int64 `json:"user_id"`
    Name string `json:"name"`
    Description string `json:"description"`
    ImageURL string `json:"image_url"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func storyExists(story *Story, fetch bool) (bool, *Story) {
    if fetch {
        var name, description, imageURL sql.NullString

        if err := db.QueryRow(GET_STORY_SQL, story.ID).Scan(
            &story.ID,
            &story.HoopID,
            &story.UserID,
            &name,
            &description,
            &imageURL,
            &story.CreatedAt,
            &story.UpdatedAt,
        ); err != nil {
            log.Println(err)
            return false, nil
        }

        story.Name = fromNullString(name)
        story.Description = fromNullString(description)
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
