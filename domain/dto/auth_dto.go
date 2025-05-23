package dto

type RequestOTPRegisterUserRequest struct {
	Email string `json:"email" validate:"required,email,max=320"`
}

type CheckOTPRegisterUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required"`
}

type RegisterUserRequest struct {
	Email    string `json:"email" validate:"required,email,max=320"`
	OTP      string `json:"otp" validate:"required"`
	Name     string `json:"name" validate:"required,min=3,max=100,ascii"`
	Password string `json:"password" validate:"required,min=8,max=72,ascii"`
}

type LoginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,ascii"`
}

type LoginResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	User         *UserResponse `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RequestOTPResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OTP         string `json:"otp" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72,ascii"`
}
