package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainUserNotFound           = errors.New("user not found")
	ErrDomainUserAlreadyExists      = errors.New("user already exists")
	ErrDomainUserEmailAlreadyExists = errors.New("user email already exists")
	ErrDomainUserPhoneAlreadyExists = errors.New("user phone already exists")
	ErrDomainUserInactive           = errors.New("user account is inactive")

	ErrDomainFriendCodeInvalid      = errors.New("friend code is invalid")
	ErrDomainAdminSelfRoleForbidden = errors.New("administrators cannot change their own role")
)

// User
var (
	ErrUserNotFound           = &AppError{Code: "USER_NOT_FOUND", Message: "user not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainUserNotFound}
	ErrUserAlreadyExists      = &AppError{Code: "USER_ALREADY_EXISTS", Message: "user already exists", HTTPStatus: http.StatusConflict, Err: ErrDomainUserAlreadyExists}
	ErrUserEmailAlreadyExists = &AppError{Code: "USER_EMAIL_ALREADY_EXISTS", Message: "email is already registered", HTTPStatus: http.StatusConflict, Err: ErrDomainUserEmailAlreadyExists}
	ErrUserPhoneAlreadyExists = &AppError{Code: "USER_PHONE_ALREADY_EXISTS", Message: "phone number is already registered", HTTPStatus: http.StatusConflict, Err: ErrDomainUserPhoneAlreadyExists}
	ErrUserInactive           = &AppError{Code: "USER_INACTIVE", Message: "user account is deactivated", HTTPStatus: http.StatusForbidden, Err: ErrDomainUserInactive}
)

var ErrFriendCodeInvalid = &AppError{
	Code:       "FRIEND_CODE_INVALID",
	Message:    "friend code must start with DR- followed by 12 letters or digits, excluding 0, 1, I and O",
	HTTPStatus: http.StatusBadRequest,
	Err:        ErrDomainFriendCodeInvalid,
}

var ErrAdminSelfRoleChange = &AppError{
	Code:       "CANNOT_CHANGE_OWN_ROLE",
	Message:    "administrators cannot change their own role",
	HTTPStatus: http.StatusForbidden,
	Err:        ErrDomainAdminSelfRoleForbidden,
}
