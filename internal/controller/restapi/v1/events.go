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

// @Summary      Create new event
// @Description  Creates new event with the number of seats and the booking duration
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        body  body      request.EventRequest    true  "Event info"
// @Success      201   {object}  response.EventResponse  "Event created"
// @Failure      400   {object}  response.Error          "Invalid request"
// @Failure      500   {object}  response.Error          "Internal error"
// @Router       /v1/events [post]
func (r *V1) createEvent(ctx *fiber.Ctx) error {
	var body request.EventRequest

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

	resp := response.EventResponse{
		ID:         e.ID,
		Name:       e.Name,
		Date:       e.Date,
		TotalSeats: e.TotalSeats,
		BookingTTL: e.BookingTTL,
	}

	return ctx.Status(fiber.StatusCreated).JSON(resp)
}

// @Summary      Create new booking
// @Description  Creates new booking by event id and user email
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        id    path      int                	   true  "Event ID"
// @Param        body  body      request.BookSeatRequest   true  "Booking info"
// @Success      201   {object}  response.BookSeatResponse "Successful booking"
// @Failure      400   {object}  response.Error     	   "Invalid request"
// @Failure      404   {object}  response.Error     	   "Event not found"
// @Failure      409   {object}  response.Error     	   "No free seats"
// @Failure      500   {object}  response.Error     	   "Internal error"
// @Router       /v1/events/{id}/book [post]
func (r *V1) book(ctx *fiber.Ctx) error {
	var body request.BookSeatRequest

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

	resp := response.BookSeatResponse{
		ID:        booking.ID,
		EventID:   booking.EventID,
		Email:     booking.Email,
		Status:    string(booking.Status),
		ExpiresAt: booking.ExpiresAt,
	}

	return ctx.Status(fiber.StatusCreated).JSON(resp)
}

// @Summary      Confirm booking
// @Description  Confirms booking by ID if it has not yet expired/confirmed already
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        id   path      int                         true  "Booking ID"
// @Success      201  {object}  response.ConfirmedBooking   "Booking confirmed"
// @Failure      400  {object}  response.Error              "Invalid request"
// @Failure      404  {object}  response.Error              "Booking not found or cancelled/confirmed"
// @Failure      500  {object}  response.Error              "Internal error"
// @Router       /v1/events/{id}/confirm [post]
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

// @Summary      Get event info
// @Description  Returns detailed information about an event, including the number of available and occupied seats
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        id   path      int                  true  "Event ID"
// @Success      200  {object}  response.EventInfo   "Event info"
// @Failure      400  {object}  response.Error       "Invalid request"
// @Failure      404  {object}  response.Error       "Event not found"
// @Failure      500  {object}  response.Error       "Internal error"
// @Router       /v1/events/{id} [get]
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

// @Summary      Get booking status
// @Description  Returns current booking status: (pending, confirmed, expired)
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        id   path      int                      true  "Booking ID"
// @Success      200  {object}  response.BookingStatus   "Current booking status"
// @Failure      400  {object}  response.Error           "Invalid request"
// @Failure      404  {object}  response.Error           "Booking not found"
// @Failure      500  {object}  response.Error           "Internal error"
// @Router       /v1/bookings/{id}/status [get]
func (r *V1) getBookingStatus(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return errorResponse(ctx, fiber.StatusBadRequest, "invalid booking id")
	}

	status, err := r.b.GetBookingStatus(ctx.UserContext(), id)
	if err != nil {
		if errors.Is(err, errs.ErrBookingNotFound) {
			return errorResponse(ctx, fiber.StatusNotFound, "booking not found")
		}
		r.logger.Error(err, "restapi - v1 - getBookingStatus")

		return errorResponse(ctx, fiber.StatusInternalServerError, "internal error")
	}

	resp := response.BookingStatus{
		Status: status,
	}

	return ctx.Status(fiber.StatusOK).JSON(resp)
}
