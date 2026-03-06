package usecase

import (
	"context"

	"github.com/andreyxaxa/Event-Booker/internal/dto"
	"github.com/andreyxaxa/Event-Booker/internal/entity"
)

type (
	BookingUseCase interface {
		CreateEvent(ctx context.Context, event entity.Event) (entity.Event, error)
		Book(ctx context.Context, eventID int64, email string) (entity.Booking, error)
		ConfirmBooking(ctx context.Context, bookingID int64) (entity.Booking, error)
		GetEventInfo(ctx context.Context, eventID int64) (entity.Event, int64, error)
		ExpireBookings(ctx context.Context, limit int64) (int, error)
		GetMessagesToSend(ctx context.Context, limit int64) ([]dto.CancelledBooking, error)
	}
)
