package errorpkg

import (
	"net/http"
)

var (
	ErrInternalServer = NewError(http.StatusInternalServerError).
		WithErrorCode("INTERNAL_SERVER_ERROR").
		WithMessage("Something went wrong in our server. Please try again later.")

	ErrCredentialsNotMatch = NewError(http.StatusUnauthorized).
		WithErrorCode("CREDENTIALS_NOT_MATCH").
		WithMessage("Credentials do not match. Please try again.")

	ErrEmailAlreadyRegistered = NewError(http.StatusConflict).
		WithErrorCode("EMAIL_ALREADY_REGISTERED").
		WithMessage("Email already registered. Please login or use another email.")

	ErrFailParseRequest = NewError(http.StatusBadRequest).
		WithErrorCode("FAIL_PARSE_REQUEST").
		WithMessage("Failed to parse request. Please check your request format.")

	ErrForbiddenRole = NewError(http.StatusForbidden).
		WithErrorCode("FORBIDDEN_ROLE").
		WithMessage("You're not allowed to access this resource.")

	ErrInvalidBearerToken = NewError(http.StatusUnauthorized).
		WithErrorCode("INVALID_BEARER_TOKEN").
		WithMessage("Your session is invalid. Please renew your session.")

	ErrInvalidOTP = NewError(http.StatusUnauthorized).
		WithErrorCode("INVALID_OTP").
		WithMessage("Invalid OTP. Please try again or request a new OTP.")

	ErrInvalidRefreshToken = NewError(http.StatusUnauthorized).
		WithErrorCode("INVALID_REFRESH_TOKEN").
		WithMessage("Session is invalid. Please login again.")

	ErrNoBearerToken = NewError(http.StatusUnauthorized).
		WithErrorCode("NO_BEARER_TOKEN").
		WithMessage("You're not logged in. Please login first.")

	ErrNotFound = NewError(http.StatusNotFound).
		WithErrorCode("NOT_FOUND").
		WithMessage("Data not found.")
)
