package repo

import (
	"context"

	"github.com/andreyxaxa/Event-Booker/internal/dto"
	"github.com/andreyxaxa/Event-Booker/internal/entity"
)

type (
	EventRepo interface {
		Create(ctx context.Context, event entity.Event) (entity.Event, error)
		TryBookSeat(ctx context.Context, eventID int64) (string, error)
		Get(ctx context.Context, eventID int64) (entity.Event, error)
		DecrementBookedSeats(ctx context.Context, eventID int64, count int64) error
	}

	BookingRepo interface {
		Create(ctx context.Context, eventID int64, email string, bookingTTL string) (entity.Booking, error)
		MarkAsConfirmed(ctx context.Context, bookingID int64) (entity.Booking, error)
		MarkCancelled(ctx context.Context, limit int64) ([]dto.CancelledBooking, error)
		GetStatus(ctx context.Context, bookingID int64) (string, error)
	}

	OutboxRepo interface {
		Create(ctx context.Context, bookings []dto.CancelledBooking) error
		FetchAndDelete(ctx context.Context, limit int64) ([]dto.CancelledBooking, error)
	}

	Transactor interface {
		WithinTranscation(ctx context.Context, f func(ctx context.Context) error) error
	}
)
