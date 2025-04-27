package dto

type RequestOTPRegisterUserRequest struct {
	Email string `json:"email" validate:"required,email,max=320"`
}
