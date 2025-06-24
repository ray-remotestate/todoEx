package models

import (
    "time"

    "github.com/google/uuid"
)

type User struct {
    ID           uuid.UUID  `db:"id" json:"id"`
    Name         string     `db:"name" json:"name"`
    Email        string     `db:"email" json:"email"`
    Password 	 string     `db:"password" json:"-"`
    CreatedAt    time.Time  `db:"created_at" json:"created_at"`
    ArchivedAt   *time.Time `db:"archived_at" json:"archived_at,omitempty"`
}

type UserSession struct {
    ID         	 uuid.UUID  `db:"id" json:"id"`
    UserID     	 uuid.UUID  `db:"user_id" json:"user_id"`
    SessionToken string     `db:"session_token" json:"session_token"`
    CreatedAt  	 time.Time  `db:"created_at" json:"created_at"`
    ExpiresAt  	 time.Time  `db:"expires_at" json:"expires_at"`
}