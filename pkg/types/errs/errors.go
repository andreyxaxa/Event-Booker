package errs

import "errors"

var (
	ErrEventNotFound            = errors.New("event not found")
	ErrBookingNotFound          = errors.New("booking not found")
	ErrBookingNotFoundOrExpired = errors.New("booking not found or expired")
	ErrNoSeatsAvailable         = errors.New("no seats available")
)
