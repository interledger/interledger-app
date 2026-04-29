package accountdeletion

import "time"

type Request struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	RequestedAt time.Time `db:"requested_at"`
}
