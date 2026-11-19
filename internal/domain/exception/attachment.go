package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainAttachmentNotFound  = errors.New("attachment not found")
	ErrDomainInvalidFileType     = errors.New("invalid file type or format")
	ErrDomainFileTooLarge        = errors.New("file size exceeds maximum allowed limit")
	ErrDomainStorageUnavailable  = errors.New("storage service unavailable")
	ErrDomainAttachmentForbidden = errors.New("permission denied for this attachment")
)

// Attachment
var (
	ErrAttachmentNotFound  = &AppError{Code: "ATTACHMENT_NOT_FOUND", Message: "attachment not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainAttachmentNotFound}
	ErrInvalidFileType     = &AppError{Code: "INVALID_FILE_TYPE", Message: "file type or format is not supported", HTTPStatus: http.StatusBadRequest, Err: ErrDomainInvalidFileType}
	ErrFileTooLarge        = &AppError{Code: "FILE_TOO_LARGE", Message: "file size exceeds maximum allowed limit", HTTPStatus: http.StatusRequestEntityTooLarge, Err: ErrDomainFileTooLarge}
	ErrStorageUnavailable  = &AppError{Code: "STORAGE_UNAVAILABLE", Message: "storage service is temporarily unavailable", HTTPStatus: http.StatusServiceUnavailable, Err: ErrDomainStorageUnavailable}
	ErrAttachmentForbidden = &AppError{Code: "ATTACHMENT_FORBIDDEN", Message: "you do not have permission to access or modify this attachment", HTTPStatus: http.StatusForbidden, Err: ErrDomainAttachmentForbidden}
)
