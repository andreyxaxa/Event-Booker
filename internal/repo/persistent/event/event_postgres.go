package event

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/andreyxaxa/Event-Booker/internal/entity"
	"github.com/andreyxaxa/Event-Booker/pkg/postgres"
	"github.com/andreyxaxa/Event-Booker/pkg/types/errs"
	"github.com/jackc/pgx/v5"
)

const (
	eventsTable = "events"

	idColumn          = "id"
	nameColumn        = "name"
	dateColumn        = "date"
	totalSeatsColumn  = "total_seats"
	bookedSeatsColumn = "booked_seats"
	bookingTTLColumn  = "booking_ttl"
)

type EventRepo struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) *EventRepo {
	return &EventRepo{pg}
}

func (r *EventRepo) Create(ctx context.Context, event entity.Event) (entity.Event, error) {
	sql, args, err := r.Builder.
		Insert(eventsTable).
		Columns(nameColumn, dateColumn, totalSeatsColumn, bookingTTLColumn).
		Values(event.Name, event.Date, event.TotalSeats, event.BookingTTL).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return entity.Event{}, fmt.Errorf("EventRepo - Create - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	err = executor.QueryRow(ctx, sql, args...).Scan(&event.ID)
	if err != nil {
		return entity.Event{}, fmt.Errorf("EventRepo - Create - executor.QueryRow.Scan: %w", err)
	}

	return event, nil
}

func (r *EventRepo) TryBookSeat(ctx context.Context, eventID int64) (string, error) {
	executor := r.GetExecutor(ctx)

	var booked, total int
	var bookingTTL string

	err := executor.QueryRow(ctx,
		"SELECT booked_seats, total_seats, booking_ttl::text FROM events WHERE id = $1 FOR UPDATE",
		eventID,
	).Scan(&booked, &total, &bookingTTL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("EventRepo - TryBookSeat: %w", errs.ErrEventNotFound)
		}
		return "", fmt.Errorf("EventRepo - TryBookSeat - QueryRow.Scan: %w", err)
	}

	if booked >= total {
		return "", fmt.Errorf("EventRepo - TryBookSeat: %w", errs.ErrNoSeatsAvailable)
	}

	sql, args, err := r.Builder.
		Update(eventsTable).
		Set(bookedSeatsColumn, squirrel.Expr(bookedSeatsColumn+" + 1")).
		Where(squirrel.Eq{idColumn: eventID}).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("EventRepo - TryBookSeat - r.Builder.ToSql: %w", err)
	}

	_, err = executor.Exec(ctx, sql, args...)
	if err != nil {
		return "", fmt.Errorf("EventRepo - TryBookSeat - executor.Exec: %w", err)
	}

	return bookingTTL, nil
}

func (r *EventRepo) Get(ctx context.Context, eventID int64) (entity.Event, error) {
	sql, args, err := r.Builder.
		Select(idColumn, nameColumn, dateColumn, totalSeatsColumn, bookedSeatsColumn, bookingTTLColumn).
		From(eventsTable).
		Where(squirrel.Eq{idColumn: eventID}).
		ToSql()
	if err != nil {
		return entity.Event{}, fmt.Errorf("EventRepo - Get - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	var event entity.Event

	err = executor.QueryRow(ctx, sql, args...).Scan(
		&event.ID,
		&event.Name,
		&event.Date,
		&event.TotalSeats,
		&event.BookedSeats,
		&event.BookingTTL,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Event{}, fmt.Errorf("EventRepo - Get: %w", errs.ErrEventNotFound)
		}

		return entity.Event{}, fmt.Errorf("EventRepo - Get - executor.QueryRow.Scan: %w", err)
	}

	return event, nil
}

func (r *EventRepo) DecrementBookedSeats(ctx context.Context, eventID int64, count int64) error {
	sql := `
		UPDATE events 
		SET booked_seats = booked_seats - $1 
		WHERE id = $2 AND booked_seats >= $1
	`

	executor := r.GetExecutor(ctx)

	_, err := executor.Exec(ctx, sql, count, eventID)
	if err != nil {
		return fmt.Errorf("EventRepo - DecrementBookedSeats - executor.Exec: %w", err)
	}

	return nil
}
