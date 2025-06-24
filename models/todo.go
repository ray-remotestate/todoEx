package models

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
    ID          uuid.UUID  `db:"id" json:"id"`
    UserID      uuid.UUID  `db:"user_id" json:"user_id"`
    Title       string     `db:"title" json:"title"`
    Description *string    `db:"description" json:"description,omitempty"`
    Status      string     `db:"status" json:"status"`
    DueDate     *time.Time `db:"due_date" json:"due_date,omitempty"`
    CreatedAt   time.Time  `db:"created_at" json:"created_at"`
    ArchivedAt  *time.Time `db:"archived_at" json:"archived_at,omitempty"`
}