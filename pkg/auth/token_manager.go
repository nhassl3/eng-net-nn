package auth

import "context"

type TokenManager interface {
	CreateToken(username, uid, role string) (string, error)
	CreateRefreshToken(username, uid, role string) (string, *Payload, error)
	VerifyToken(ctx context.Context, token string) (*Payload, error)
	GetTTL() int
}
