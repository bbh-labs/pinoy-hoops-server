package main

import (
    "time"
)

type Activity struct {
    UserID int64 `json:"user_id"`
    Type int64 `json:"type"`
    HoopID int64 `json:"hoop_id,omitempty"`
    StoryID int64 `json:"story_id,omitempty"`
    CommentID int64 `json:"comment_id,omitempty"`
    Data map[string]interface{} `json:"data"`
    CreatedAt time.Time `json:"created_at"`
}

const (
    ACTIVITY_POST_HOOP = 1
    ACTIVITY_POST_STORY = 2
    ACTIVITY_POST_COMMENT_HOOP = 101
    ACTIVITY_POST_COMMENT_STORY = 102
    ACTIVITY_POST_LIKE_HOOP = 201
    ACTIVITY_POST_LIKE_STORY = 201
)
