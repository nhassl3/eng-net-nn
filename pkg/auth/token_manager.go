package auth

type TokenManager interface {
	CreateToken(username, uid, role string) (string, error)
	CreateRefreshToken(username, uid, role string) (string, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}
