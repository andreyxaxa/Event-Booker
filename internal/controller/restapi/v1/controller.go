package v1

import (
	"github.com/andreyxaxa/Event-Booker/internal/usecase"
	"github.com/andreyxaxa/Event-Booker/pkg/logger"
	"github.com/go-playground/validator/v10"
)

type V1 struct {
	b      usecase.BookingUseCase
	logger logger.Interface
	v      *validator.Validate
}
