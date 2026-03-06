package restapi

import (
	v1 "github.com/andreyxaxa/Event-Booker/internal/controller/restapi/v1"
	"github.com/andreyxaxa/Event-Booker/internal/usecase"
	"github.com/andreyxaxa/Event-Booker/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

func NewRouter(app *fiber.App, b usecase.BookingUseCase, l logger.Interface) {
	apiV1Group := app.Group("/v1")
	{
		v1.NewEventsRoutes(apiV1Group, b, l)
	}
}
