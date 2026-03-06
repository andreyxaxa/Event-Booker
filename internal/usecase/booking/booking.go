package booking

import (
	"context"
	"fmt"

	"github.com/andreyxaxa/Event-Booker/internal/dto"
	"github.com/andreyxaxa/Event-Booker/internal/entity"
	"github.com/andreyxaxa/Event-Booker/internal/repo"
)

type BookingUseCase struct {
	eventRepo   repo.EventRepo
	bookingRepo repo.BookingRepo
	outboxRepo  repo.OutboxRepo
	transactor  repo.Transactor
}

func New(e repo.EventRepo, b repo.BookingRepo, o repo.OutboxRepo, t repo.Transactor) *BookingUseCase {
	return &BookingUseCase{
		eventRepo:   e,
		bookingRepo: b,
		outboxRepo:  o,
		transactor:  t,
	}
}

func (uc *BookingUseCase) CreateEvent(ctx context.Context, event entity.Event) (entity.Event, error) {
	e, err := uc.eventRepo.Create(ctx, event)
	if err != nil {
		return entity.Event{}, fmt.Errorf("BookingUseCase - CreateEvent - uc.eventRepo.Create: %w", err)
	}

	return e, nil
}

func (uc *BookingUseCase) Book(ctx context.Context, eventID int64, email string) (entity.Booking, error) {
	var booking entity.Booking

	// в транзакции
	err := uc.transactor.WithinTranscation(ctx, func(ctx context.Context) error {
		bookingTTL, err := uc.eventRepo.TryBookSeat(ctx, eventID)
		if err != nil {
			return fmt.Errorf("BookingUseCase - Book - uc.eventRepo.TryBookSeat: %w", err)
		}

		booking, err = uc.bookingRepo.Create(ctx, eventID, email, bookingTTL)
		if err != nil {
			return fmt.Errorf("BookingUseCase - Book - uc.bookingRepo.Create: %w", err)
		}

		return nil
	})

	if err != nil {
		return booking, fmt.Errorf("BookingUseCase - Book - uc.transactor.WithinTranscation: %w", err)
	}

	return booking, nil
}

func (uc *BookingUseCase) ConfirmBooking(ctx context.Context, bookingID int64) (entity.Booking, error) {
	booking, err := uc.bookingRepo.MarkAsConfirmed(ctx, bookingID)
	if err != nil {
		return entity.Booking{}, fmt.Errorf("BookingUseCase - ConfirmBooking - uc.bookingRepo.MarkAsConfirmed: %w", err)
	}

	return booking, nil
}

func (uc *BookingUseCase) GetEventInfo(ctx context.Context, eventID int64) (entity.Event, int64, error) {
	event, err := uc.eventRepo.Get(ctx, eventID)
	if err != nil {
		return entity.Event{}, 0, fmt.Errorf("BookingUseCase - GetEventInfo - uc.eventRepo.Get: %w", err)
	}

	freeSeats := event.TotalSeats - event.BookedSeats

	return event, freeSeats, nil
}

func (uc *BookingUseCase) ExpireBookings(ctx context.Context, limit int64) (int, error) {
	cancelledCount := 0

	// в транзакции
	err := uc.transactor.WithinTranscation(ctx, func(ctx context.Context) error {
		// у тех, что истекли, меняем статус
		// получаем емейлы и мероприятия, в которых нужно освободить забронированные места
		cancelled, err := uc.bookingRepo.MarkCancelled(ctx, limit)
		if err != nil {
			return fmt.Errorf("BookingUseCase - HandleExpiredBookings - uc.bookingRepo.MarkCancelled: %w", err)
		}

		cancelledCount = len(cancelled)

		// считаем, сколько истекших по каждому eventID
		// TODO: возможно в этом же цикле куда-то емейлы переложить
		counts := make(map[int64]int64)
		for _, b := range cancelled {
			counts[b.EventID]++
		}

		// для каждого мероприятия освобождаем места
		for eventID, count := range counts {
			err = uc.eventRepo.DecrementBookedSeats(ctx, eventID, count)
			if err != nil {
				return fmt.Errorf("BookingUseCase - HandleExpiredBookings - uc.eventRepo.DecrementBookedSeats: %w", err)
			}
		}

		// отправляем в аутбокс для дальнейшей рассылки
		err = uc.outboxRepo.Create(ctx, cancelled)
		if err != nil {
			return fmt.Errorf("BookingUseCase - HandleExpiredBookings - uc.outboxRepo.Create: %w", err)
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("BookingUseCase - HandleExpiredBookings - uc.transactor.WithinTranscation: %w", err)
	}

	return cancelledCount, nil
}

func (uc *BookingUseCase) GetMessagesToSend(ctx context.Context, limit int64) ([]dto.CancelledBooking, error) {
	cb, err := uc.outboxRepo.FetchAndDelete(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("BookingUseCase - GetMessagesToSend - uc.outboxRepo.FetchAndDelete: %w", err)
	}

	return cb, nil
}
