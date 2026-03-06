package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/andreyxaxa/Event-Booker/internal/dto"
	"github.com/andreyxaxa/Event-Booker/internal/entity"
	"github.com/andreyxaxa/Event-Booker/pkg/postgres"
	"github.com/andreyxaxa/Event-Booker/pkg/types/errs"
	"github.com/jackc/pgx/v5"
)

// TODO: возможно удалить
const (
	bookingsTable   = "bookings"
	idColumn        = "id"
	eventIDColumn   = "event_id"
	emailColumn     = "email"
	statusColumn    = "status"
	expiresAtColumn = "expires_at"
	createdAtColumn = "created_at"
)

type BookingRepo struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) *BookingRepo {
	return &BookingRepo{pg}
}

func (r *BookingRepo) Create(ctx context.Context, eventID int64, email string, bookingTTL string) (entity.Booking, error) {
	executor := r.GetExecutor(ctx)

	var booking entity.Booking

	err := executor.QueryRow(ctx, `
		INSERT INTO bookings (event_id, email, expires_at)
		VALUES ($1, $2, now() + $3::interval)
		RETURNING id, event_id, email, status, expires_at, created_at
	`, eventID, email, bookingTTL).Scan(
		&booking.ID,
		&booking.EventID,
		&booking.Email,
		&booking.Status,
		&booking.ExpiresAt,
		&booking.CreatedAt,
	)
	if err != nil {
		return entity.Booking{}, fmt.Errorf("BookingRepo - Create - executor.QueryRow.Scan: %w", err)
	}

	return booking, nil
}

func (r *BookingRepo) MarkAsConfirmed(ctx context.Context, bookingID int64) (entity.Booking, error) {
	sql, args, err := r.Builder.
		Update(bookingsTable).
		Set(statusColumn, string(entity.Confirmed)).
		Where(squirrel.And{
			squirrel.Eq{idColumn: bookingID},
			squirrel.Eq{statusColumn: string(entity.Pending)},
			squirrel.Gt{expiresAtColumn: time.Now()},
		}).
		Suffix("RETURNING id, event_id, status").
		ToSql()
	if err != nil {
		return entity.Booking{}, fmt.Errorf("BookingRepo - MarkAsConfirmed - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	var b entity.Booking

	err = executor.QueryRow(ctx, sql, args...).Scan(&b.ID, &b.EventID, &b.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Booking{}, fmt.Errorf("BookingRepo - MarkAsConfirmed: %w", errs.ErrBookingNotFoundOrExpired)
		}

		return entity.Booking{}, fmt.Errorf("BookingRepo - MarkAsConfirmed - executor.QueryRow.Scan: %w", err)
	}

	return b, nil
}

func (r *BookingRepo) MarkCancelled(ctx context.Context, limit int64) ([]dto.CancelledBooking, error) {
	sql := `
	UPDATE bookings
	SET status = 'cancelled'
	WHERE id IN (
		SELECT id FROM bookings
		WHERE status = 'pending' AND expires_at <= NOW()
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	)
	RETURNING email, event_id;
	`

	executor := r.GetExecutor(ctx)

	rows, err := executor.Query(ctx, sql, limit)
	if err != nil {
		return nil, fmt.Errorf("BookingRepo - MarkCancelled - executor.Query: %w", err)
	}
	defer rows.Close()

	cancelledBookings := make([]dto.CancelledBooking, 0, limit)

	for rows.Next() {
		var cancelledBooking dto.CancelledBooking
		err = rows.Scan(&cancelledBooking.Email, &cancelledBooking.EventID)
		if err != nil {
			return nil, fmt.Errorf("BookingRepo - MarkCancelled - rows.Scan: %w", err)
		}
		cancelledBookings = append(cancelledBookings, cancelledBooking)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("BookingRepo - MarkCancelled - rows.Err: %w", err)
	}

	return cancelledBookings, nil
}

func (r *BookingRepo) GetStatus(ctx context.Context, bookingID int64) (string, error) {
	sql, args, err := r.Builder.
		Select(statusColumn).
		From(bookingsTable).
		Where(squirrel.Eq{idColumn: bookingID}).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("BookingRepo - GetStatus - r.Builder.ToSql: %w", err)
	}

	executor := r.GetExecutor(ctx)

	var status string

	err = executor.QueryRow(ctx, sql, args...).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("BookingRepo - GetStatus: %w", errs.ErrBookingNotFound)
		}
		return "", fmt.Errorf("BookingRepo - GetStatus - QueryRow.Scan: %w", err)
	}

	return status, nil
}
