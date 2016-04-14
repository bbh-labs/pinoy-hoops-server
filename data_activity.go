package main

import (
    "database/sql"
    "time"
)

type Activity struct {
    UserID int64 `json:"user_id"`
    Type int64 `json:"type"`
    HoopID sql.NullInt64 `json:"hoop_id"`
    StoryID sql.NullInt64 `json:"story_id"`
    Data map[string]interface{} `json:"data"`
    CreatedAt time.Time `json:"created_at"`
}

const (
    ACTIVITY_POST_HOOP = 1
    ACTIVITY_POST_STORY = 2
)
