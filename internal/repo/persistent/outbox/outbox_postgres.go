package outbox

import (
	"context"
	"fmt"

	"github.com/andreyxaxa/Event-Booker/internal/dto"
	"github.com/andreyxaxa/Event-Booker/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type OutboxRepo struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) *OutboxRepo {
	return &OutboxRepo{pg}
}

func (r *OutboxRepo) Create(ctx context.Context, bookings []dto.CancelledBooking) error {
	if len(bookings) == 0 {
		return nil
	}

	inputRows := [][]any{}
	for _, b := range bookings {
		inputRows = append(inputRows, []any{b.EventID, b.Email})
	}

	executor := r.GetExecutor(ctx)

	_, err := executor.CopyFrom(
		ctx,
		pgx.Identifier{"outbox"},
		[]string{"event_id", "email"},
		pgx.CopyFromRows(inputRows),
	)
	if err != nil {
		return fmt.Errorf("OutboxRepo - Create - executor.CopyFrom: %w", err)
	}

	return nil
}

func (r *OutboxRepo) FetchAndDelete(ctx context.Context, limit int64) ([]dto.CancelledBooking, error) {
	sql := `
	DELETE FROM outbox
	WHERE id IN (
		SELECT id FROM outbox
		ORDER BY id ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	)
	RETURNING event_id, email;
	`

	executor := r.GetExecutor(ctx)

	rows, err := executor.Query(ctx, sql, limit)
	if err != nil {
		return nil, fmt.Errorf("OutboxRepo - FetchAndDelete - executor.Query: %w", err)
	}
	defer rows.Close()

	var cbs []dto.CancelledBooking
	for rows.Next() {
		var cb dto.CancelledBooking
		err = rows.Scan(&cb.EventID, &cb.Email)
		if err != nil {
			return nil, fmt.Errorf("OutboxRepo - FetchAndDelete - rows.Scan: %w", err)
		}
		cbs = append(cbs, cb)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("OutboxRepo - FetchAndDelete - rows.Err: %w", err)
	}

	return cbs, nil
}
