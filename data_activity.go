package main

import (
	"database/sql"
	"time"
)

const (
	ACTIVITY_POST_HOOP          = 1
	ACTIVITY_POST_STORY         = 2
	ACTIVITY_POST_COMMENT_HOOP  = 101
	ACTIVITY_POST_COMMENT_STORY = 102
	ACTIVITY_POST_LIKE_HOOP     = 201
	ACTIVITY_POST_LIKE_STORY    = 202
)

type Activity struct {
	UserID    int64                  `json:"user_id"`
	Type      int64                  `json:"type"`
	HoopID    int64                  `json:"hoop_id,omitempty"`
	StoryID   int64                  `json:"story_id,omitempty"`
	CommentID int64                  `json:"comment_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

func (a *Activity) fetchData() {
	a.Data = make(map[string]interface{})

	if ok, user := userExists(&User{ID: a.UserID}, true); ok {
		a.Data["user"] = *user
	}

	switch a.Type {
	case ACTIVITY_POST_HOOP:
		if ok, hoop := hoopExists(&Hoop{ID: a.HoopID}, true); ok {
			a.Data["hoop"] = *hoop
		}
	case ACTIVITY_POST_STORY:
		if ok, story := storyExists(&Story{ID: a.StoryID}, true); ok {
			a.Data["story"] = *story
		}
	case ACTIVITY_POST_LIKE_HOOP:
		if ok, hoop := hoopExists(&Hoop{ID: a.HoopID}, true); ok {
			a.Data["hoop"] = *hoop
		}
	case ACTIVITY_POST_LIKE_STORY:
		if ok, story := storyExists(&Story{ID: a.StoryID}, true); ok {
			a.Data["story"] = *story
		}
	case ACTIVITY_POST_COMMENT_HOOP:
		if ok, hoop := hoopExists(&Hoop{ID: a.HoopID}, true); ok {
			a.Data["hoop"] = *hoop
		}
	case ACTIVITY_POST_COMMENT_STORY:
		if ok, story := storyExists(&Story{ID: a.StoryID}, true); ok {
			a.Data["story"] = *story
		}
	}
}

func getActivities(userID int64) ([]Activity, error) {
	var activities []Activity

	rows, err := db.Query(GET_ACTIVITIES_SQL, userID)
	if err != nil {
		return nil, err
	}

	var hoopID, storyID sql.NullInt64

	for rows.Next() {
		var activity Activity

		if err := rows.Scan(
			&activity.UserID,
			&activity.Type,
			&hoopID,
			&storyID,
			&activity.CreatedAt,
		); err != nil {
			return nil, err
		}

		activity.HoopID = fromNullInt64(hoopID)
		activity.StoryID = fromNullInt64(storyID)
		activity.fetchData()

		activities = append(activities, activity)
	}

	return activities, nil
}
