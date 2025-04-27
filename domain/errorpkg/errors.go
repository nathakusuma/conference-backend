package errorpkg

import (
	"net/http"
)

var (
	ErrInternalServer = NewError(http.StatusInternalServerError).
				WithErrorCode("INTERNAL_SERVER_ERROR").
				WithMessage("Something went wrong in our server. Please try again later.")

	ErrFailParseRequest = NewError(http.StatusBadRequest).
				WithErrorCode("FAIL_PARSE_REQUEST").
				WithMessage("Failed to parse request. Please check your request format.")

	ErrNotFound = NewError(http.StatusNotFound).
			WithErrorCode("NOT_FOUND").
			WithMessage("Data not found.")

	ErrEmailAlreadyRegistered = NewError(http.StatusConflict).
					WithErrorCode("EMAIL_ALREADY_REGISTERED").
					WithMessage("Email already registered. Please login or use another email.")
)
