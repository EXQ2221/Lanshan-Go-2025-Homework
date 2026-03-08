package errcode

import "errors"

var (
	ErrBadRequest      = errors.New("bad request")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrConflict        = errors.New("conflict")
	ErrNotFound        = errors.New("not found")
	ErrInternal        = errors.New("internal server error")

	ErrUsernameIncorrect   = errors.New("username incorrect")
	ErrPasswordIncorrect   = errors.New("password incorrect")
	ErrHasFollowed         = errors.New("has followed")
	ErrHasNotFollowed      = errors.New("has not followed")
	ErrInvalidListType     = errors.New("invalid list type")
)
