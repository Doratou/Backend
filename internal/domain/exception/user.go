package exception

import (
	"errors"
	"net/http"
)

var ErrDomainFriendCodeInvalid = errors.New("friend code is invalid")
var ErrDomainAdminSelfRoleForbidden = errors.New("administrators cannot change their own role")

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
