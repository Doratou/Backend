package exception

import (
	"errors"
	"net/http"
)

var ErrDomainFriendCodeInvalid = errors.New("friend code is invalid")

var ErrFriendCodeInvalid = &AppError{
	Code:       "FRIEND_CODE_INVALID",
	Message:    "friend code must start with DR- followed by 12 letters or digits, excluding 0, 1, I and O",
	HTTPStatus: http.StatusBadRequest,
	Err:        ErrDomainFriendCodeInvalid,
}
