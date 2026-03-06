package v1

import (
	"github.com/andreyxaxa/Event-Booker/internal/usecase"
	"github.com/andreyxaxa/Event-Booker/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func NewEventsRoutes(apiV1Group fiber.Router, b usecase.BookingUseCase, l logger.Interface) {
	r := &V1{b: b, logger: l, v: validator.New(validator.WithRequiredStructEnabled())}

	eventsGroup := apiV1Group.Group("/events")
	bookingsGroup := apiV1Group.Group("/bookings")

	{
		// API
		// /events
		eventsGroup.Post("/", r.createEvent)
		eventsGroup.Post("/:id/book", r.book)
		eventsGroup.Post("/:id/confirm", r.confirmBook)
		eventsGroup.Get("/:id", r.getEventInfo)

		// /bookings
		bookingsGroup.Get("/:id", r.getBookingStatus)
	}
}
