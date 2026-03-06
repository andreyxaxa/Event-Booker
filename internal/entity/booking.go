package entity

import "time"

type BookingStatus string

const (
	Pending   BookingStatus = "pending"
	Confirmed BookingStatus = "confirmed"
	Cancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID        int64         `json:"id"`
	EventID   int64         `json:"event_id"`
	Email     string        `json:"email"`
	Status    BookingStatus `json:"status"`
	ExpiresAt time.Time     `json:"expires_at"`
	CreatedAt time.Time     `json:"created_at"`
}
