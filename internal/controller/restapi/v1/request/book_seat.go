package request

type BookSeatRequest struct {
	Email string `json:"email" validate:"required,email"`
}
