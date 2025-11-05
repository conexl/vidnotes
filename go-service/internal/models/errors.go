// models/errors.go
package models

import "errors"

var (
	ErrSessionCreateFailed = errors.New("session create failed")
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionUpdateFailed = errors.New("session update failed")
	ErrSessionDeleteFailed = errors.New("session delete failed")
	ErrInvalidMessageRole  = errors.New("invalid message role")

	ErrUserAlreadyExists            = errors.New("user already exists")
	ErrUserCreateFailed             = errors.New("user create failed")
	ErrUserNotFound                 = errors.New("user not found")
	ErrUserUpdateFailed             = errors.New("user update failed")
	ErrUserDeleteFailed             = errors.New("user delete failed")
	ErrInvalidPassword              = errors.New("invalid password")
	ErrMonthlyAnalysesLimitExceeded = errors.New("monthly analyses limit exceeded")
	ErrInvalidSubscription          = errors.New("invalid subscription")

	ErrVideoCreateFailed = errors.New("video create failed")
	ErrVideoNotFound     = errors.New("video not found")
	ErrVideoUpdateFailed = errors.New("video update failed")
	ErrVideoDeleteFailed = errors.New("video delete failed")

	ErrVideoResultCreateFailed = errors.New("video result create failed")
	ErrVideoResultNotFound     = errors.New("video result not found")
	ErrVideoResultUpdateFailed = errors.New("video result update failed")

	ErrAIServiceUnavailable = errors.New("AI service unavailable")
	ErrInvalidAIRequest     = errors.New("invalid AI request")
	ErrContextTooLong       = errors.New("context too long")
)
