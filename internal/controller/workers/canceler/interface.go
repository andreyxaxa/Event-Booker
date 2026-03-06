package canceler

import "context"

type BookingsExpirer interface {
	ExpireBookings(ctx context.Context, limit int64) (int, error)
}
