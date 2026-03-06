package notifier

import (
	"context"

	"github.com/andreyxaxa/Event-Booker/internal/dto"
)

type (
	MessagesGetter interface {
		GetMessagesToSend(ctx context.Context, limit int64) ([]dto.CancelledBooking, error)
	}

	MessageSender interface {
		Send(cb dto.CancelledBooking) error
	}
)
