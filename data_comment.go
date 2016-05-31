package main

import (
    "database/sql"
    "log"
	"time"
)

type Comment struct {
	ID        int64                  `json:"id"`
	UserID    int64                  `json:"user_id"`
	HoopID    int64                  `json:"hoop_id,omitempty"`
	StoryID   int64                  `json:"story_id,omitempty"`
	User      User                   `json:"user"`
	Text      string                 `json:"text"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
    Data      map[string]interface{} `json:"data,omitempty"`
}

func insertHoopComment(userID, hoopID int64, text string) error {
    // Start Transaction
    tx, err := db.Begin()
    if err != nil {
        return err
    }

    // Insert Comment
    if _, err = tx.Exec(INSERT_HOOP_COMMENT_SQL, userID, text, hoopID); err != nil {
        return err
    }

    // Insert Activity
    if _, err = db.Exec(INSERT_HOOP_COMMENT_ACTIVITY_SQL, userID, ACTIVITY_POST_COMMENT_HOOP, hoopID); err != nil {
        return err
    }

    // End Transaction
    if err := tx.Commit(); err != nil {
        return err
    }

    return nil
}

func insertStoryComment(userID, storyID int64, text string) error {
    // Start Transaction
    tx, err := db.Begin()
    if err != nil {
        return err
    }

    // Insert Comment
    if _, err = tx.Exec(INSERT_STORY_COMMENT_SQL, userID, text, storyID); err != nil {
        return err
    }

    // Insert Activity
    if _, err = db.Exec(INSERT_STORY_COMMENT_ACTIVITY_SQL, userID, ACTIVITY_POST_COMMENT_STORY, storyID); err != nil {
        return err
    }

    // End Transaction
    if err := tx.Commit(); err != nil {
        return err
    }

    return nil
}

func getHoopComments(hoopID int64) ([]Comment, error) {
    var comments []Comment

    rows, err := db.Query(GET_HOOP_COMMENTS_SQL, hoopID)
    if err != nil {
        return nil, err
    }

    var text sql.NullString

    for rows.Next() {
        var comment Comment

        if err := rows.Scan(
            &comment.ID,
            &comment.UserID,
            &text,
            &comment.HoopID,
            &comment.CreatedAt,
            &comment.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        comment.Text = fromNullString(text)

        user, err := getUserByID(comment.UserID)
        if err != nil {
            log.Println(err)
            continue
        }
        comment.Data = make(map[string]interface{})
        comment.Data["user"] = user

        comments = append(comments, comment)
    }

    return comments, nil
}

func getStoryComments(storyID int64) ([]Comment, error) {
    var comments []Comment

    rows, err := db.Query(GET_STORY_COMMENTS_SQL, storyID)
    if err != nil {
        return nil, err
    }

    var text sql.NullString

    for rows.Next() {
        var comment Comment

        if err := rows.Scan(
            &comment.ID,
            &comment.UserID,
            &text,
            &comment.StoryID,
            &comment.CreatedAt,
            &comment.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        comment.Text = fromNullString(text)

        if comment.User, err = getUserByID(comment.UserID); err != nil {
            return nil, err
        }

        comments = append(comments, comment)
    }

    return comments, nil
}
