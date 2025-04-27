package dto

type RequestOTPRegisterUserRequest struct {
	Email string `json:"email" validate:"required,email,max=320"`
}

type CheckOTPRegisterUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required"`
}
