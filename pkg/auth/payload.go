package auth

import "time"

type Payload struct {
	JTI       string    `json:"jti"`
	Username  string    `json:"username"`
	UID       string    `json:"uid"`
	Role      string    `json:"role"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func (p *Payload) Valid() error {
	if time.Now().After(p.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
