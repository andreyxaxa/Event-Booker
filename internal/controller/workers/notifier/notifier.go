package notifier

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andreyxaxa/Event-Booker/internal/dto"
	"github.com/andreyxaxa/Event-Booker/pkg/logger"
)

type Notifier struct {
	mg MessagesGetter
	ms MessageSender
	l  logger.Interface

	batchSize          int64
	workers            int
	pollInterval       time.Duration
	getMessagesTimeout time.Duration
	// TODO: sendTimeout

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	started atomic.Bool
}

func New(
	mg MessagesGetter,
	ms MessageSender,
	l logger.Interface,
	batchSize int64,
	workers int,
	pollInterval time.Duration,
	getMessagesTimeout time.Duration,
) *Notifier {
	return &Notifier{
		mg:                 mg,
		ms:                 ms,
		l:                  l,
		batchSize:          batchSize,
		workers:            workers,
		pollInterval:       pollInterval,
		getMessagesTimeout: getMessagesTimeout,
	}
}

func (n *Notifier) Start(ctx context.Context) error {
	if !n.started.CompareAndSwap(false, true) {
		return fmt.Errorf("Notifier - Start - worker already started")
	}

	n.ctx, n.cancel = context.WithCancel(ctx)

	tasks := make(chan dto.CancelledBooking, n.workers*2)

	for i := 0; i < n.workers; i++ {
		n.wg.Add(1)
		go n.sender(n.ctx, tasks)
	}

	n.wg.Add(1)
	go n.fetcher(n.ctx, tasks)

	return nil
}

func (n *Notifier) sender(ctx context.Context, tasks <-chan dto.CancelledBooking) {
	defer n.wg.Done()

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					n.l.Error(fmt.Errorf("panic %v", r), "Notifier - sender - panic")
				}
			}()
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-tasks:
				if !ok {
					return
				}
				err := n.ms.Send(msg)
				if err != nil {
					n.l.Error(err, "Notifier - sender - n.ms.Send")
					return
				}
			}
		}()

		if ctx.Err() != nil {
			return
		}
	}
}

func (n *Notifier) fetcher(ctx context.Context, tasks chan<- dto.CancelledBooking) {
	defer n.wg.Done()

	ticker := time.NewTicker(n.pollInterval)
	defer ticker.Stop()

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					n.l.Error(fmt.Errorf("panic %v", r), "Notifier - fetcher - panic")
				}
			}()
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for {
					gmCtx, gmCancel := context.WithTimeout(ctx, n.getMessagesTimeout)
					msgs, err := n.mg.GetMessagesToSend(gmCtx, n.batchSize)
					gmCancel()
					if err != nil {
						n.l.Error(err, "Notifier - fetcher - n.m.GetMessagesToSend")
						return
					}
					for _, msg := range msgs {
						select {
						case <-ctx.Done():
							return
						case tasks <- msg:
						}
					}

					// если вернули меньше лимита - значит в базе мало - работаем по тикеру
					if int64(len(msgs)) < n.batchSize {
						break
					}

					// если вернули по лимиту - значит в базе много - не ждем тикера - крутимся в беск. цикле
					n.l.Info("Notifier - fetcher: batch full, processing next one")
				}
			}
		}()

		if ctx.Err() != nil {
			return
		}
	}
}

func (n *Notifier) Shutdown(ctx context.Context) error {
	if !n.started.Load() {
		return nil
	}

	if n.cancel != nil {
		n.cancel()
	}

	done := make(chan struct{})

	go func() {
		n.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
