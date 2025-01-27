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

	ErrEndTimeBeforeStart = NewError(http.StatusUnprocessableEntity).
		WithErrorCode("END_TIME_BEFORE_START").
		WithMessage("End time is before start time. Please use correct time.")

	ErrFailParseRequest = NewError(http.StatusBadRequest).
		WithErrorCode("FAIL_PARSE_REQUEST").
		WithMessage("Failed to parse request. Please check your request format.")

	ErrForbiddenRole = NewError(http.StatusForbidden).
		WithErrorCode("FORBIDDEN_ROLE").
		WithMessage("You're not allowed to access this resource.")

	ErrForbiddenUser = NewError(http.StatusForbidden).
		WithErrorCode("FORBIDDEN_USER").
		WithMessage("You're not allowed to access this resource.")

	ErrInvalidBearerToken = NewError(http.StatusUnauthorized).
		WithErrorCode("INVALID_BEARER_TOKEN").
		WithMessage("Your auth session is invalid. Please renew your auth session.")

	ErrInvalidOTP = NewError(http.StatusUnauthorized).
		WithErrorCode("INVALID_OTP").
		WithMessage("Invalid OTP. Please try again or request a new OTP.")

	ErrInvalidRefreshToken = NewError(http.StatusUnauthorized).
		WithErrorCode("INVALID_REFRESH_TOKEN").
		WithMessage("Auth session is invalid. Please login again.")

	ErrNoBearerToken = NewError(http.StatusUnauthorized).
		WithErrorCode("NO_BEARER_TOKEN").
		WithMessage("You're not logged in. Please login first.")

	ErrNotFound = NewError(http.StatusNotFound).
		WithErrorCode("NOT_FOUND").
		WithMessage("Data not found.")

	ErrTimeAlreadyPassed = NewError(http.StatusUnprocessableEntity).
		WithErrorCode("TIME_ALREADY_PASSED").
		WithMessage("Time has already passed. Please use future time.")

	ErrTimeWindowConflict = NewError(http.StatusConflict).
		WithErrorCode("TIME_WINDOW_CONFLICT").
		WithMessage("There's already a conference in the same time window. Please choose another time window.")

	ErrUpdateApprovedTimeWindow = NewError(http.StatusUnprocessableEntity).
		WithErrorCode("UPDATE_APPROVED_TIME_WINDOW").
		WithMessage("You're not allowed to update an approved conference time window.")

	ErrUpdatePastConference = NewError(http.StatusUnprocessableEntity).
		WithErrorCode("UPDATE_PAST_CONFERENCE").
		WithMessage("You're not allowed to update a past conference.")

	ErrUpdateRejectedConference = NewError(http.StatusUnprocessableEntity).
		WithErrorCode("UPDATE_REJECTED_CONFERENCE").
		WithMessage("You're not allowed to update a rejected conference.")

	ErrUpdatePastConferenceStatus = NewError(http.StatusUnprocessableEntity).
		WithErrorCode("UPDATE_PAST_CONFERENCE_STATUS").
		WithMessage("You're only allowed to change past conference status from pending to rejected.")

	ErrUserHasActiveProposal = NewError(http.StatusConflict).
		WithErrorCode("USER_HAS_ACTIVE_PROPOSAL").
		WithMessage("You already have an active proposal. Please wait until it's accepted or delete it.")
)
