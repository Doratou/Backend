package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainBadRequest = errors.New("bad request")
	ErrDomainInternal   = errors.New("internal server error")
)

// General
var (
	ErrBadRequest = &AppError{Code: "BAD_REQUEST", Message: "bad request payload or parameters", HTTPStatus: http.StatusBadRequest, Err: ErrDomainBadRequest}
	ErrInternal   = &AppError{Code: "INTERNAL_SERVER_ERROR", Message: "an internal server error occurred", HTTPStatus: http.StatusInternalServerError, Err: ErrDomainInternal}
)
