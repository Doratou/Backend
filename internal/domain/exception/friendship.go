package exception

import (
	"errors"
	"net/http"
)

var (
	ErrDomainFriendshipNotFound         = errors.New("friendship not found")
	ErrDomainFriendshipAlreadyExists    = errors.New("a friendship already exists between these users")
	ErrDomainFriendRequestAlreadyExists = errors.New("a pending friend request already exists")
	ErrDomainAlreadyFriends             = errors.New("users are already friends")
	ErrDomainSelfFriendRequest          = errors.New("cannot send a friend request to yourself")
	ErrDomainFriendshipForbidden        = errors.New("you are not allowed to perform this friendship operation")
	ErrDomainFriendshipInvalidStatus    = errors.New("friendship status does not allow this operation")
	ErrDomainFriendRequestCooldown      = errors.New("friend request cooldown is active")
)

// Friendship
var (
	ErrFriendshipNotFound         = &AppError{Code: "FRIENDSHIP_NOT_FOUND", Message: "friendship not found", HTTPStatus: http.StatusNotFound, Err: ErrDomainFriendshipNotFound}
	ErrFriendshipAlreadyExists    = &AppError{Code: "FRIENDSHIP_ALREADY_EXISTS", Message: "a friendship already exists between these users", HTTPStatus: http.StatusConflict, Err: ErrDomainFriendshipAlreadyExists}
	ErrFriendRequestAlreadyExists = &AppError{Code: "FRIEND_REQUEST_ALREADY_EXISTS", Message: "a pending friend request already exists", HTTPStatus: http.StatusConflict, Err: ErrDomainFriendRequestAlreadyExists}
	ErrAlreadyFriends             = &AppError{Code: "ALREADY_FRIENDS", Message: "users are already friends", HTTPStatus: http.StatusConflict, Err: ErrDomainAlreadyFriends}
	ErrSelfFriendRequest          = &AppError{Code: "SELF_FRIEND_REQUEST", Message: "cannot send a friend request to yourself", HTTPStatus: http.StatusBadRequest, Err: ErrDomainSelfFriendRequest}
	ErrFriendshipForbidden        = &AppError{Code: "FRIENDSHIP_FORBIDDEN", Message: "you are not allowed to perform this friendship operation", HTTPStatus: http.StatusForbidden, Err: ErrDomainFriendshipForbidden}
	ErrFriendshipInvalidStatus    = &AppError{Code: "FRIENDSHIP_INVALID_STATUS", Message: "friendship status does not allow this operation", HTTPStatus: http.StatusConflict, Err: ErrDomainFriendshipInvalidStatus}
	ErrFriendRequestCooldown      = &AppError{Code: "FRIEND_REQUEST_COOLDOWN", Message: "please wait before sending another friend request to this user", HTTPStatus: http.StatusTooManyRequests, Err: ErrDomainFriendRequestCooldown}
)
