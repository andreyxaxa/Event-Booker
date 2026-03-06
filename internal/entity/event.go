package entity

import "time"

type Event struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	Date        time.Time     `json:"date"`
	TotalSeats  int64         `json:"total_seats"`
	BookedSeats int64         `json:"booked_seats"`
	BookingTTL  time.Duration `json:"booking_ttl"`
}
