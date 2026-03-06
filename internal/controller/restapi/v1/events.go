package v1

import (
	"errors"
	"strconv"

	"github.com/andreyxaxa/Event-Booker/internal/controller/restapi/v1/request"
	"github.com/andreyxaxa/Event-Booker/internal/controller/restapi/v1/response"
	"github.com/andreyxaxa/Event-Booker/internal/entity"
	"github.com/andreyxaxa/Event-Booker/pkg/types/errs"
	"github.com/gofiber/fiber/v2"
)

func (r *V1) createEvent(ctx *fiber.Ctx) error {
	var body request.Event

	err := ctx.BodyParser(&body)
	if err != nil {
		return errorResponse(ctx, fiber.StatusBadRequest, "invalid request body")
	}

	event := entity.Event{
		Name:       body.Name,
		Date:       body.Date,
		TotalSeats: body.TotalSeats,
		BookingTTL: body.BookingTTL.Duration,
	}

	e, err := r.b.CreateEvent(ctx.UserContext(), event)
	if err != nil {
		r.logger.Error(err, "restapi - v1 - createEvent")

		return errorResponse(ctx, fiber.StatusInternalServerError, "internal error")
	}

	resp := response.Event{
		ID:         e.ID,
		Name:       e.Name,
		Date:       e.Date,
		TotalSeats: e.TotalSeats,
		BookingTTL: e.BookingTTL,
	}

	return ctx.Status(fiber.StatusCreated).JSON(resp)
}

func (r *V1) book(ctx *fiber.Ctx) error {
	var body request.BookSeat

	err := ctx.BodyParser(&body)
	if err != nil {
		return errorResponse(ctx, fiber.StatusBadRequest, "invalid request body")
	}

	idStr := ctx.Params("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return errorResponse(ctx, fiber.StatusBadRequest, "invalid event id")
	}

	booking, err := r.b.Book(ctx.UserContext(), id, body.Email)
	if err != nil {
		if errors.Is(err, errs.ErrEventNotFound) {
			return errorResponse(ctx, fiber.StatusNotFound, "event not found")
		} else if errors.Is(err, errs.ErrNoSeatsAvailable) {
			return errorResponse(ctx, fiber.StatusConflict, "all seats for this event are taken")
		}
		r.logger.Error(err, "restapi - v1 - book")

		return errorResponse(ctx, fiber.StatusInternalServerError, "internal error")
	}

	resp := response.BookSeat{
		ID:        booking.ID,
		EventID:   booking.EventID,
		Email:     booking.Email,
		Status:    string(booking.Status),
		ExpiresAt: booking.ExpiresAt,
	}

	return ctx.Status(fiber.StatusCreated).JSON(resp)
}

func (r *V1) confirmBook(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return errorResponse(ctx, fiber.StatusBadRequest, "invalid booking id")
	}

	booking, err := r.b.ConfirmBooking(ctx.UserContext(), id)
	if err != nil {
		if errors.Is(err, errs.ErrBookingNotFoundOrExpired) {
			return errorResponse(ctx, fiber.StatusNotFound, "booking not found or expired/confirmed")
		}
		r.logger.Error(err, "restapi - v1 - confirmBook")

		return errorResponse(ctx, fiber.StatusInternalServerError, "internal error")
	}

	resp := response.ConfirmedBooking{
		ID:      booking.ID,
		EventID: booking.EventID,
		Status:  string(booking.Status),
	}

	return ctx.Status(fiber.StatusCreated).JSON(resp)
}

func (r *V1) getEventInfo(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return errorResponse(ctx, fiber.StatusBadRequest, "invalid booking id")
	}

	event, freeSeats, err := r.b.GetEventInfo(ctx.UserContext(), id)
	if err != nil {
		if errors.Is(err, errs.ErrEventNotFound) {
			return errorResponse(ctx, fiber.StatusNotFound, "event not found")
		}
		r.logger.Error(err, "restapi - v1 - getEventInfo")

		return errorResponse(ctx, fiber.StatusInternalServerError, "internal error")
	}

	resp := response.EventInfo{
		ID:          event.ID,
		Name:        event.Name,
		Date:        event.Date,
		TotalSeats:  event.TotalSeats,
		BookedSeats: event.BookedSeats,
		FreeSeats:   freeSeats,
		BookingTTL:  event.BookingTTL,
	}

	return ctx.Status(fiber.StatusOK).JSON(resp)
}
