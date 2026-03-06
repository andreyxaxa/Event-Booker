package response

import "time"

type BookSeatResponse struct {
	ID        int64     `json:"id"`
	EventID   int64     `json:"event_id"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
}
