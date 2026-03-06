package request

import (
	"encoding/json"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	d.Duration = dur

	return nil
}

type EventRequest struct {
	Name       string    `json:"name"`
	Date       time.Time `json:"date"`
	TotalSeats int64     `json:"total_seats"`
	BookingTTL Duration  `json:"booking_ttl" swaggertype:"string" example:"1h"`
}
