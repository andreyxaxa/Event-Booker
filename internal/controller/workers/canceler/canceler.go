package canceler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andreyxaxa/Event-Booker/pkg/logger"
)

type BookingCanceler struct {
	b BookingsExpirer
	l logger.Interface

	batchSize       int64
	pollInterval    time.Duration
	expireOpTimeout time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	started atomic.Bool
}

func New(b BookingsExpirer, l logger.Interface, batchSize int64, pollInterval time.Duration, expireOpTimeout time.Duration) *BookingCanceler {
	return &BookingCanceler{
		b:               b,
		l:               l,
		batchSize:       batchSize,
		pollInterval:    pollInterval,
		expireOpTimeout: expireOpTimeout,
	}
}

func (c *BookingCanceler) Start(ctx context.Context) error {
	if !c.started.CompareAndSwap(false, true) {
		return fmt.Errorf("BookingCanceler - Start - worker already started")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)

	ticker := time.NewTicker(c.pollInterval)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer ticker.Stop()
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						c.l.Error(fmt.Errorf("panic %v", r), "BookingCanceler - Start - panic")
					}
				}()

				select {
				case <-c.ctx.Done():
					return
				case <-ticker.C:
					c.process(c.ctx)
				}
			}()

			if c.ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

func (c *BookingCanceler) process(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		expCtx, expCancel := context.WithTimeout(ctx, c.expireOpTimeout)
		cnt, err := c.b.ExpireBookings(expCtx, c.batchSize)
		expCancel()
		if err != nil {
			c.l.Error(err, "BookingCanceler - process - c.b.ExpireBookings")
			return
		}

		// если вернули меньше лимита - значит в базе мало - работаем по тикеру
		if int64(cnt) < c.batchSize {
			break
		}

		// если вернули по лимиту - значит в базе много - не ждем тикера - крутимся в беск. цикле
		c.l.Info("BookingCanceler: batch full, processing next one")
	}
}

func (c *BookingCanceler) Shutdown(ctx context.Context) error {
	if !c.started.Load() {
		return nil
	}

	if c.cancel != nil {
		c.cancel()
	}

	done := make(chan struct{})

	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
