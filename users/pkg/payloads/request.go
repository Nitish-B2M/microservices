package payloads

type UserUpdateRequest struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Gender    string `json:"gender"`
}

type RegisterUser struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" validate:"required"`
}
