package entity

import (
	"time"
)

type User struct {
	ID        int       `db:"id"`
	Balance   float64   `db:"balance"`
	CreatedAt time.Time `db:"created_at"`
}
