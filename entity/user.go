package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `db:"id"`
	Balance   float64   `db:"balance"`
	CreatedAt time.Time `db:"created_at"`
}
