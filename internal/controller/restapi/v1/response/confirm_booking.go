package response

type ConfirmedBooking struct {
	ID      int64  `json:"id"`
	EventID int64  `json:"event_id"`
	Status  string `json:"status"`
}
