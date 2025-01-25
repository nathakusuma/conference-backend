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
)
