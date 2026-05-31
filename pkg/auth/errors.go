package auth

import "errors"

var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is invalid")
	ErrTokenRevoked = errors.New("token has been revoked")
)

// IsAny checks if error given in args in token realization errors slice
func IsAny(err error) bool {
	for _, t := range []error{ErrExpiredToken, ErrInvalidToken, ErrTokenRevoked} {
		if errors.Is(err, t) {
			return true
		}
	}
	return false
}
