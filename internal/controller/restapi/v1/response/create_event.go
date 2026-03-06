package response

import "time"

type EventResponse struct {
	ID         int64         `json:"id"`
	Name       string        `json:"name"`
	Date       time.Time     `json:"date"`
	TotalSeats int64         `json:"total_seats"`
	BookingTTL time.Duration `json:"booking_ttl" swaggertype:"string" example:"1h"`
}
