package user

type GetUserRequest struct {
	UserID uint64 `json:"user_id" validate:"required"`
}

type UpdateProfileRequest struct {
	UserID      uint64 `json:"-"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}
