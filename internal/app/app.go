package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/andreyxaxa/Event-Booker/config"
	"github.com/andreyxaxa/Event-Booker/internal/controller/restapi"
	"github.com/andreyxaxa/Event-Booker/internal/controller/workers/canceler"
	"github.com/andreyxaxa/Event-Booker/internal/controller/workers/notifier"
	"github.com/andreyxaxa/Event-Booker/internal/infrastructure/mailsender"
	"github.com/andreyxaxa/Event-Booker/internal/repo/persistent/booking"
	"github.com/andreyxaxa/Event-Booker/internal/repo/persistent/event"
	"github.com/andreyxaxa/Event-Booker/internal/repo/persistent/outbox"
	bookingUC "github.com/andreyxaxa/Event-Booker/internal/usecase/booking"
	"github.com/andreyxaxa/Event-Booker/pkg/httpserver"
	"github.com/andreyxaxa/Event-Booker/pkg/logger"
	"github.com/andreyxaxa/Event-Booker/pkg/postgres"
	"github.com/andreyxaxa/Event-Booker/pkg/smtpsender"
)

func Run(cfg *config.Config) {
	// Context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Logger
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, l, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// Use-Case
	bookingUseCase := bookingUC.New(
		event.New(pg),
		booking.New(pg),
		outbox.New(pg),
		pg,
	)

	// Canceler Worker
	cancelerWorker := canceler.New(
		bookingUseCase,
		l,
		cfg.CancelerWorker.BatchSize,
		cfg.CancelerWorker.PollInterval,
		cfg.CancelerWorker.ExpireOpTimeout,
	)

	// Mail Sender
	mailSender := mailsender.New(smtpsender.New(
		smtpsender.Username(cfg.SMTP.Username),
		smtpsender.Password(cfg.SMTP.Password),
		smtpsender.Host(cfg.SMTP.Host),
		smtpsender.Port(cfg.SMTP.Port),
	))

	// Notifier Worker
	notifierWorker := notifier.New(
		bookingUseCase,
		mailSender,
		l,
		cfg.NotifierWorker.BatchSize,
		cfg.NotifierWorker.Workers,
		cfg.NotifierWorker.PollInterval,
		cfg.NotifierWorker.GetMessagesTimeout,
	)

	// HTTP Server
	httpServer := httpserver.New(l, httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
	restapi.NewRouter(httpServer.App, bookingUseCase, l)

	// Start Components
	err = cancelerWorker.Start(ctx)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - cancelerWorker.Start: %w", err))
	}

	err = notifierWorker.Start(ctx)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - notifierWorker.Start: %w", err))
	}

	httpServer.Start()

	// Waiting Signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: %s", s.String())
	case err := <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}

	cwShutdownCtx, cwShutdownCancel := context.WithTimeout(ctx, cfg.CancelerWorker.ShutdownTimeout)
	defer cwShutdownCancel()
	err = cancelerWorker.Shutdown(cwShutdownCtx)
	if err != nil {
		l.Error(fmt.Errorf("app - Run - cancelerWorker.Shutdown: %w", err))
	}

	nwShutdownCtx, nwShutdownCancel := context.WithTimeout(ctx, cfg.NotifierWorker.ShutdownTimeout)
	defer nwShutdownCancel()
	err = notifierWorker.Shutdown(nwShutdownCtx)
	if err != nil {
		l.Error(fmt.Errorf("app - Run - notifierWorker.Shutdown: %w", err))
	}
}
