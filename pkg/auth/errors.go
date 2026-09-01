package auth

import "errors"

var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is invalid")
	ErrTokenRevoked = errors.New("token has been revoked")
)
