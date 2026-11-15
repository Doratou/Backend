package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainReportNotFound        = errors.New("report not found")
	ErrDomainReportTargetRequired  = errors.New("reported user id or reported group id must be provided")
	ErrDomainReportSelfForbidden   = errors.New("you cannot report yourself")
	ErrDomainInvalidReportReason   = errors.New("invalid report reason")
	ErrDomainInvalidReportStatus   = errors.New("invalid report status")
	ErrDomainReportAlreadyResolved = errors.New("report has already been resolved")
	ErrDomainForbiddenRole         = errors.New("forbidden: insufficient permissions")
)

// Admin Report & Moderation
var (
	ErrReportNotFound        = &AppError{Code: "REPORT_NOT_FOUND", Message: "report not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainReportNotFound}
	ErrReportTargetRequired  = &AppError{Code: "REPORT_TARGET_REQUIRED", Message: "at least one of reported user id or group id is required", HTTPStatus: http.StatusBadRequest, Err: ErrDomainReportTargetRequired}
	ErrReportSelfForbidden   = &AppError{Code: "REPORT_SELF_FORBIDDEN", Message: "you cannot report yourself", HTTPStatus: http.StatusBadRequest, Err: ErrDomainReportSelfForbidden}
	ErrInvalidReportReason   = &AppError{Code: "INVALID_REPORT_REASON", Message: "invalid report reason provided", HTTPStatus: http.StatusBadRequest, Err: ErrDomainInvalidReportReason}
	ErrInvalidReportStatus   = &AppError{Code: "INVALID_REPORT_STATUS", Message: "invalid report status provided", HTTPStatus: http.StatusBadRequest, Err: ErrDomainInvalidReportStatus}
	ErrReportAlreadyResolved = &AppError{Code: "REPORT_ALREADY_RESOLVED", Message: "report has already been reviewed and resolved", HTTPStatus: http.StatusBadRequest, Err: ErrDomainReportAlreadyResolved}
	ErrForbiddenRole         = &AppError{Code: "FORBIDDEN_ROLE", Message: "you do not have permission to access this resource", HTTPStatus: http.StatusForbidden, Err: ErrDomainForbiddenRole}
)
