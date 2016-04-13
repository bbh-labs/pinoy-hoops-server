package main

import (
    "time"
)

type Story struct {
    ID int64 `json:"id"`
    UserID int64 `json:"user_id"`
    Name string `json:"name"`
    Description string `json:"description"`
    ImageURL string `json:"image_url"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
