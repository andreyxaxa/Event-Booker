package restapi

import (
	"github.com/andreyxaxa/Event-Booker/config"
	_ "github.com/andreyxaxa/Event-Booker/docs" // Swagger docs.
	v1 "github.com/andreyxaxa/Event-Booker/internal/controller/restapi/v1"
	"github.com/andreyxaxa/Event-Booker/internal/usecase"
	"github.com/andreyxaxa/Event-Booker/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// @title Event Booker
// @version 1.0.0
// @host localhost:8080
// @BasePath /v1
func NewRouter(app *fiber.App, cfg *config.Config, b usecase.BookingUseCase, l logger.Interface) {
	// Swagger
	if cfg.Swagger.Enabled {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// Routers
	apiV1Group := app.Group("/v1")
	{
		v1.NewEventsRoutes(apiV1Group, b, l)
	}
}
