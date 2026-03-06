package request

type BookSeat struct {
	Email string `json:"email" validate:"required,email"`
}
