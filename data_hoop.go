package main

import (
    "time"
)

type Hoop struct {
    ID int64 `json:"id"`
    UserID int64 `json:"user_id"`
    Name string `json:"name"`
    Description string `json:"description"`
    Latitude float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
